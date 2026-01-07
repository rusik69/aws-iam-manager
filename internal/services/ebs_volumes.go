package services

import (
	"fmt"
	"sync"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ============================================================================
// EBS VOLUME MANAGEMENT
// ============================================================================

func (s *AWSService) ListEBSVolumes() ([]models.EBSVolume, error) {
	const cacheKey = "ebs-volumes"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if volumes, ok := cached.([]models.EBSVolume); ok {
			return volumes, nil
		}
	}

	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %v", err)
	}

	// Filter accessible accounts
	var accessibleAccounts []models.Account
	for _, account := range accounts {
		if account.Accessible {
			accessibleAccounts = append(accessibleAccounts, account)
		}
	}

	if len(accessibleAccounts) == 0 {
		return []models.EBSVolume{}, nil
	}

	// Channel to collect results from goroutines
	type accountResult struct {
		volumes   []models.EBSVolume
		err       error
		accountID string
	}

	resultChan := make(chan accountResult, len(accessibleAccounts))
	var wg sync.WaitGroup

	// Process each account in parallel
	for _, account := range accessibleAccounts {
		wg.Add(1)
		go func(acc models.Account) {
			defer wg.Done()

			volumes, err := s.ListEBSVolumesByAccount(acc.ID)
			resultChan <- accountResult{
				volumes:   volumes,
				err:       err,
				accountID: acc.ID,
			}
		}(account)
	}

	// Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allVolumes []models.EBSVolume
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get volumes for account %s: %v\n", result.accountID, result.err)
			continue
		}
		allVolumes = append(allVolumes, result.volumes...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allVolumes, s.cacheTTL)

	return allVolumes, nil
}

// ListEBSVolumesByAccount returns all EBS volumes for a specific account
func (s *AWSService) ListEBSVolumesByAccount(accountID string) ([]models.EBSVolume, error) {
	cacheKey := fmt.Sprintf("ebs-volumes:%s", accountID)

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if volumes, ok := cached.([]models.EBSVolume); ok {
			return volumes, nil
		}
	}

	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		fmt.Printf("[WARNING] Cannot access account %s, skipping volume listing: %v\n", accountID, err)
		return []models.EBSVolume{}, nil
	}

	// Get account name
	accounts, _ := s.ListAccounts()
	accountName := accountID
	for _, acc := range accounts {
		if acc.ID == accountID {
			accountName = acc.Name
			break
		}
	}

	// Get all regions
	ec2Client := ec2.New(sess)
	regionsResult, err := ec2Client.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %v", err)
	}

	if len(regionsResult.Regions) == 0 {
		return []models.EBSVolume{}, nil
	}

	var allVolumes []models.EBSVolume
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process each region in parallel
	for _, region := range regionsResult.Regions {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()

			regionSess := sess.Copy(&aws.Config{Region: aws.String(r)})
			ec2Client := ec2.New(regionSess)

			// List volumes in this region
			input := &ec2.DescribeVolumesInput{}

			result, err := ec2Client.DescribeVolumes(input)
			if err != nil {
				fmt.Printf("[WARNING] Failed to list volumes in %s for account %s: %v\n", r, accountID, err)
				return
			}

			var regionVolumes []models.EBSVolume
			for _, vol := range result.Volumes {
				// Extract volume name from tags
				volumeName := ""
				for _, tag := range vol.Tags {
					if *tag.Key == "Name" {
						volumeName = *tag.Value
						break
					}
				}

				// Convert tags
				var tags []models.Tag
				for _, tag := range vol.Tags {
					tags = append(tags, models.Tag{
						Key:   aws.StringValue(tag.Key),
						Value: aws.StringValue(tag.Value),
					})
				}

				// Convert attachments
				var attachments []models.VolumeAttachment
				for _, att := range vol.Attachments {
					attachment := models.VolumeAttachment{
						InstanceID: aws.StringValue(att.InstanceId),
						Device:     aws.StringValue(att.Device),
						State:      aws.StringValue(att.State),
					}
					if att.AttachTime != nil {
						attachment.AttachTime = *att.AttachTime
					}
					attachments = append(attachments, attachment)
				}

				volume := models.EBSVolume{
					VolumeID:         aws.StringValue(vol.VolumeId),
					Name:             volumeName,
					AccountID:        accountID,
					AccountName:      accountName,
					Region:           r,
					Size:             aws.Int64Value(vol.Size),
					VolumeType:       aws.StringValue(vol.VolumeType),
					State:            aws.StringValue(vol.State),
					AvailabilityZone: aws.StringValue(vol.AvailabilityZone),
					Encrypted:        aws.BoolValue(vol.Encrypted),
					SnapshotID:       aws.StringValue(vol.SnapshotId),
					Attachments:      attachments,
					Tags:             tags,
				}

				if vol.CreateTime != nil {
					volume.CreateTime = *vol.CreateTime
				}
				if vol.Iops != nil {
					volume.IOPS = *vol.Iops
				}
				if vol.Throughput != nil {
					volume.Throughput = *vol.Throughput
				}

				regionVolumes = append(regionVolumes, volume)
			}

			mu.Lock()
			allVolumes = append(allVolumes, regionVolumes...)
			mu.Unlock()
		}(*region.RegionName)
	}

	wg.Wait()

	// Cache the result
	s.cache.Set(cacheKey, allVolumes, s.cacheTTL)

	return allVolumes, nil
}

// DetachEBSVolume detaches an EBS volume from all instances
func (s *AWSService) DetachEBSVolume(accountID, region, volumeID string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
	ec2Client := ec2.New(regionSess)

	// Get volume details to find attachments
	volResult, err := ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volumeID)},
	})
	if err != nil {
		return fmt.Errorf("failed to describe volume %s: %v", volumeID, err)
	}

	if len(volResult.Volumes) == 0 {
		return fmt.Errorf("volume %s not found", volumeID)
	}

	vol := volResult.Volumes[0]
	if len(vol.Attachments) == 0 {
		return fmt.Errorf("volume %s is not attached to any instances", volumeID)
	}

	// Detach from all instances
	for _, attachment := range vol.Attachments {
		_, err = ec2Client.DetachVolume(&ec2.DetachVolumeInput{
			VolumeId:   aws.String(volumeID),
			InstanceId: attachment.InstanceId,
			Device:     attachment.Device,
		})
		if err != nil {
			return fmt.Errorf("failed to detach volume %s from instance %s: %v", volumeID, *attachment.InstanceId, err)
		}
	}

	// Invalidate cache to refresh volume state
	s.InvalidateEBSVolumesCache()

	return nil
}

func (s *AWSService) DeleteEBSVolume(accountID, region, volumeID string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
	ec2Client := ec2.New(regionSess)

	// Check if the volume is in-use before attempting deletion
	volResult, err := ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volumeID)},
	})
	if err != nil {
		return fmt.Errorf("failed to describe volume %s: %v", volumeID, err)
	}

	if len(volResult.Volumes) == 0 {
		return fmt.Errorf("volume %s not found", volumeID)
	}

	vol := volResult.Volumes[0]
	if len(vol.Attachments) > 0 {
		return fmt.Errorf("volume %s is still attached to instance(s). Please detach first", volumeID)
	}

	_, err = ec2Client.DeleteVolume(&ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume %s: %v", volumeID, err)
	}

	// Update cache - remove the deleted volume
	s.updateEBSVolumeCache(accountID, volumeID)

	return nil
}

// updateEBSVolumeCache removes a deleted volume from the cache
func (s *AWSService) updateEBSVolumeCache(accountID, volumeID string) {
	// Update account-specific cache
	cacheKey := fmt.Sprintf("ebs-volumes:%s", accountID)
	if cached, found := s.cache.Get(cacheKey); found {
		if volumes, ok := cached.([]models.EBSVolume); ok {
			var updated []models.EBSVolume
			for _, vol := range volumes {
				if vol.VolumeID != volumeID {
					updated = append(updated, vol)
				}
			}
			s.cache.Set(cacheKey, updated, s.cacheTTL)
		}
	}

	// Update global cache
	if cached, found := s.cache.Get("ebs-volumes"); found {
		if volumes, ok := cached.([]models.EBSVolume); ok {
			var updated []models.EBSVolume
			for _, vol := range volumes {
				if vol.VolumeID != volumeID {
					updated = append(updated, vol)
				}
			}
			s.cache.Set("ebs-volumes", updated, s.cacheTTL)
		}
	}
}

// InvalidateEBSVolumesCache invalidates the EBS volumes cache
func (s *AWSService) InvalidateEBSVolumesCache() {
	s.cache.Delete("ebs-volumes")
	s.cache.DeletePattern("ebs-volumes:")
}

// getAccessKeysForUser retrieves access keys for a user with last used information
