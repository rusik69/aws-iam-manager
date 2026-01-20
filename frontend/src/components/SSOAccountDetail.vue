<template>
  <div class="sso-account-detail-container">
    <div class="page-header">
      <div class="header-content">
        <button @click="goBack" class="btn btn-secondary">
          ← Back
        </button>
        <div class="header-title">
          <h1>{{ account?.account_name || account?.account_id || 'Account Assignments' }}</h1>
          <p v-if="account">{{ account.account_id }}</p>
        </div>
        <div class="header-actions">
          <button @click="downloadJSON" class="btn btn-secondary" :disabled="!account">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19,12V19H5V12H3V19A2,2 0 0,0 5,21H19A2,2 0 0,0 21,19V12M13,12.67L15.59,10.08L17,11.5L12,16.5L7,11.5L8.41,10.08L11,12.67V3H13V12.67Z"/>
            </svg>
            Download JSON
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Loading account assignments...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <div class="error-icon">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
        </svg>
      </div>
      <h3>Failed to Load Account Assignments</h3>
      <p>{{ error }}</p>
      <button @click="loadData" class="btn btn-primary">Try Again</button>
    </div>

    <div v-else-if="account" class="main-content">
      <div class="account-info-card">
        <h2>Account Information</h2>
        <div class="info-grid">
          <div class="info-item">
            <label>Account Name:</label>
            <span>{{ account.account_name || '-' }}</span>
          </div>
          <div class="info-item">
            <label>Account ID:</label>
            <span>{{ account.account_id }}</span>
          </div>
          <div class="info-item">
            <label>Total Assignments:</label>
            <span>{{ assignments.length }}</span>
          </div>
        </div>
      </div>

      <div class="assignments-card">
        <h2>SSO Assignments ({{ assignments.length }})</h2>
        <div v-if="assignments.length === 0" class="empty-state">
          <p>No SSO assignments found for this account</p>
        </div>
        <div v-else class="assignments-list">
          <div v-for="assignment in assignments" :key="`${assignment.principal_id}-${assignment.permission_set_arn}`" class="assignment-item">
            <div class="assignment-header">
              <div class="assignment-info">
                <span class="badge" :class="{ 'user-badge': assignment.principal_type === 'USER', 'group-badge': assignment.principal_type === 'GROUP' }">
                  {{ assignment.principal_type }}
                </span>
                <span v-if="assignment.principal_type === 'USER'" @click="viewUser(assignment.principal_id)" class="link principal-name">
                  {{ assignment.principal_name || assignment.principal_id }}
                </span>
                <span v-else-if="assignment.principal_type === 'GROUP'" @click="viewGroup(assignment.principal_id)" class="link principal-name">
                  {{ assignment.principal_name || assignment.principal_id }}
                </span>
                <span v-else class="principal-name">{{ assignment.principal_name || assignment.principal_id }}</span>
                <span class="permission-set">{{ assignment.permission_set_name || assignment.permission_set_arn }}</span>
              </div>
              <button 
                v-if="assignment.principal_type === 'GROUP'"
                @click="toggleGroupMembers(assignment.principal_id)"
                class="btn btn-sm btn-secondary"
              >
                <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
                  <path v-if="expandedGroups[assignment.principal_id]" d="M7.41,15.41L12,10.83L16.59,15.41L18,14L12,8L6,14L7.41,15.41Z"/>
                  <path v-else d="M7.41,8.58L12,13.17L16.59,8.58L18,10L12,16L6,10L7.41,8.58Z"/>
                </svg>
                {{ expandedGroups[assignment.principal_id] ? 'Hide' : 'Show' }} Members 
                <span v-if="groupMembers[assignment.principal_id]">({{ groupMembers[assignment.principal_id].length }})</span>
              </button>
            </div>
            <div v-if="assignment.principal_type === 'GROUP' && expandedGroups[assignment.principal_id]" class="group-members">
              <div class="members-header">Group Members:</div>
              <div v-if="groupMembers[assignment.principal_id] === undefined" class="loading-members">
                Loading members...
              </div>
              <div v-else-if="groupMembers[assignment.principal_id] && groupMembers[assignment.principal_id].length > 0" class="members-list">
                <span 
                  v-for="member in groupMembers[assignment.principal_id]" 
                  :key="member.member_id"
                  @click="viewUser(member.member_id)"
                  class="member-badge link"
                >
                  {{ member.display_name || member.user_name || member.member_id }}
                </span>
              </div>
              <div v-else class="no-members">
                No members in this group
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'SSOAccountDetail',
  props: {
    accountId: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      account: null,
      assignments: [],
      groupMembers: {},
      expandedGroups: {},
      loading: true,
      error: null
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      this.error = null
      try {
        // Initialize account object first
        this.account = {
          account_id: this.accountId,
          account_name: null
        }
        
        // Try to get account name from accounts list
        try {
          const accountsResponse = await axios.get('/api/accounts')
          if (Array.isArray(accountsResponse.data)) {
            const accountInfo = accountsResponse.data.find(acc => acc.id === this.accountId)
            if (accountInfo) {
              this.account.account_name = accountInfo.name
            }
          }
        } catch (accountsErr) {
          console.warn('Could not fetch account name:', accountsErr)
        }
        
        // Get assignments
        const response = await axios.get(`/api/sso/accounts/${this.accountId}/assignments`)
        this.assignments = Array.isArray(response.data) ? response.data : []
        
        // Update account name from assignments if available
        if (this.assignments.length > 0 && this.assignments[0].account_name) {
          this.account.account_name = this.assignments[0].account_name
        }
        
        // Load group members for all group assignments
        await this.loadGroupMembers()
      } catch (err) {
        console.error('Error loading account assignments:', err)
        if (err.response?.status === 404 || err.response?.status === 503) {
          const errorMsg = err.response?.data?.error || 'SSO service is not available'
          const details = err.response?.data?.details || 'Ensure IAM Identity Center is enabled and the service has the required permissions.'
          if (errorMsg.includes('Identity Center') || errorMsg.includes('permissions')) {
            this.error = errorMsg
          } else {
            this.error = `${errorMsg}. ${details}`
          }
        } else {
          this.error = err.response?.data?.error || err.response?.data?.details || err.message || 'Failed to load account assignments'
        }
      } finally {
        this.loading = false
      }
    },
    async loadGroupMembers() {
      // Get unique group IDs from assignments
      const groupIds = [...new Set(
        this.assignments
          .filter(a => a.principal_type === 'GROUP')
          .map(a => a.principal_id)
      )]
      
      // Load members for each group
      for (const groupId of groupIds) {
        try {
          const response = await axios.get(`/api/sso/groups/${groupId}/members`)
          this.groupMembers[groupId] = Array.isArray(response.data) ? response.data : []
        } catch (err) {
          console.warn(`Failed to load members for group ${groupId}:`, err)
          this.groupMembers[groupId] = []
        }
      }
    },
    toggleGroupMembers(groupId) {
      this.expandedGroups[groupId] = !this.expandedGroups[groupId]
    },
    viewUser(userId) {
      this.$router.push(`/aws/sso/users/${userId}`)
    },
    viewGroup(groupId) {
      this.$router.push(`/aws/sso/groups/${groupId}`)
    },
    goBack() {
      // Preserve query params when going back to maintain filter state
      const query = { ...this.$route.query }
      this.$router.push({
        path: '/aws/sso/account-assignments',
        query
      })
    },
    downloadJSON() {
      const data = {
        account: this.account,
        assignments: this.assignments,
        groupMembers: this.groupMembers
      }
      const dataStr = JSON.stringify(data, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `sso-account-${this.accountId}-assignments-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    }
  }
}
</script>

<style scoped>
.sso-account-detail-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.header-title h1 {
  margin: 0;
  font-size: 24px;
  color: var(--color-text-primary);
}

.header-title p {
  margin: 5px 0 0 0;
  color: var(--color-text-secondary);
}

.account-info-card, .assignments-card {
  background: var(--color-bg-primary);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid var(--color-border);
}

.account-info-card h2, .assignments-card h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
  color: var(--color-text-primary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.info-item label {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.info-item span {
  color: var(--color-text-primary);
  font-size: 16px;
}

.assignments-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.assignment-item {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
}

.assignment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.assignment-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  flex-wrap: wrap;
}

.principal-name {
  font-weight: 500;
  font-size: 16px;
}

.permission-set {
  color: var(--color-text-secondary);
  font-size: 14px;
  padding: 4px 8px;
  background: var(--color-bg-tertiary);
  border-radius: 4px;
}

.group-members {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--color-border);
}

.members-header {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 12px;
  text-transform: uppercase;
  margin-bottom: 8px;
}

.members-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.member-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #10b981;
  color: white;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
}

.member-badge:hover {
  background: #059669;
}

.no-members {
  color: var(--color-text-tertiary);
  font-size: 12px;
  padding: 8px 0;
}

.loading-members {
  color: var(--color-text-secondary);
  font-size: 12px;
  padding: 8px 0;
  font-style: italic;
}

.badge {
  display: inline-block;
  padding: 4px 8px;
  background: #3b82f6;
  color: white;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.badge.user-badge {
  background: #10b981;
}

.badge.group-badge {
  background: #8b5cf6;
}

.link {
  color: var(--color-btn-primary);
  cursor: pointer;
  text-decoration: underline;
}

.link:hover {
  color: var(--color-btn-primary-hover);
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.loading-container, .error-container {
  text-align: center;
  padding: 40px;
  color: var(--color-text-primary);
}

.loading-container p {
  color: var(--color-text-secondary);
}

.error-container h3 {
  color: var(--color-text-primary);
}

.error-container p {
  color: var(--color-text-secondary);
}

.error-icon {
  color: var(--color-danger);
  margin-bottom: 10px;
}

.loading-spinner {
  border: 4px solid var(--color-border);
  border-top: 4px solid var(--color-btn-primary);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
