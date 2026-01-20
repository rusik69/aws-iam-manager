<template>
  <div class="sso-groups-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <div class="header-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M16,4C18.21,4 20,5.79 20,8C20,10.21 18.21,12 16,12C13.79,12 12,10.21 12,8C12,5.79 13.79,4 16,4M16,6C14.9,6 14,6.9 14,8C14,9.1 14.9,10 16,10C17.1,10 18,9.1 18,8C18,6.9 17.1,6 16,6M16,13C18.67,13 24,14.33 24,17V20H8V17C8,14.33 13.33,13 16,13M16,14.9C13.03,14.9 9.9,16.36 9.9,17V18.1H22.1V17C22.1,16.36 18.97,14.9 16,14.9M12.5,11.5C13.85,11.5 15,12.65 15,14C15,15.35 13.85,16.5 12.5,16.5C11.15,16.5 10,15.35 10,14C10,12.65 11.15,11.5 12.5,11.5M5.5,6C7.43,6 9,7.57 9,9.5C9,11.43 7.43,13 5.5,13C3.57,13 2,11.43 2,9.5C2,7.57 3.57,6 5.5,6M5.5,7.5C4.67,7.5 4,8.17 4,9C4,9.83 4.67,10.5 5.5,10.5C6.33,10.5 7,9.83 7,9C7,8.17 6.33,7.5 5.5,7.5M5.5,14C7.43,14 11,15.07 11,17V20H0V17C0,15.07 3.57,14 5.5,14M5.5,15.65C3.63,15.65 1.9,16.5 1.9,17V18.1H9.1V17C9.1,16.5 7.37,15.65 5.5,15.65Z"/>
            </svg>
          </div>
          <div class="title-content">
            <h1>SSO Groups</h1>
            <p>{{ ssoGroups.length }} groups</p>
          </div>
        </div>
        <div class="header-actions">
          <button @click="downloadJSON" class="btn btn-secondary" :disabled="loading || ssoGroups.length === 0">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19,12V19H5V12H3V19A2,2 0 0,0 5,21H19A2,2 0 0,0 21,19V12M13,12.67L15.59,10.08L17,11.5L12,16.5L7,11.5L8.41,10.08L11,12.67V3H13V12.67Z"/>
            </svg>
            Download JSON
          </button>
          <button @click="refreshData" class="btn btn-secondary" :disabled="loading">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
            </svg>
            {{ loading ? 'Refreshing...' : 'Refresh' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Loading SSO groups...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <div class="error-icon">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
        </svg>
      </div>
      <h3>Failed to Load SSO Groups</h3>
      <p>{{ error }}</p>
      <button @click="refreshData" class="btn btn-primary">Try Again</button>
    </div>

    <div v-else class="main-content">
      <div class="search-box">
        <svg class="search-icon" viewBox="0 0 24 24" fill="currentColor">
          <path d="M9.5,3A6.5,6.5 0 0,1 16,9.5C16,11.11 15.41,12.59 14.44,13.73L14.71,14H15.5L20.5,19L19,20.5L14,15.5V14.71L13.73,14.44C12.59,15.41 11.11,16 9.5,16A6.5,6.5 0 0,1 3,9.5A6.5,6.5 0 0,1 9.5,3M9.5,5C7,5 5,7 5,9.5C5,12 7,14 9.5,14C12,14 14,12 14,9.5C14,7 12,5 9.5,5Z"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search groups..."
          class="search-input"
        />
      </div>

      <div class="groups-table-container">
        <table class="groups-table">
          <thead>
            <tr>
              <th>Display Name</th>
              <th>Description</th>
              <th>Account Assignments</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="group in filteredGroups" :key="group.group_id">
              <td>{{ group.display_name }}</td>
              <td>{{ group.description || '-' }}</td>
              <td>
                <span class="badge">{{ group.account_assignments?.length || 0 }} accounts</span>
              </td>
              <td>
                <button 
                  @click="viewGroup(group.group_id)"
                  class="btn btn-sm btn-primary"
                >
                  View Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="filteredGroups.length === 0" class="empty-results">
          <h4>No groups found</h4>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'SSOGroups',
  data() {
    return {
      ssoGroups: [],
      loading: true,
      error: null,
      searchQuery: ''
    }
  },
  computed: {
    filteredGroups() {
      if (!this.searchQuery) return this.ssoGroups
      const query = this.searchQuery.toLowerCase()
      return this.ssoGroups.filter(group => 
        group.display_name?.toLowerCase().includes(query) ||
        group.description?.toLowerCase().includes(query)
      )
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
        const response = await axios.get('/api/sso/groups')
        const groups = response.data
        // Load assignments for each group
        for (const group of groups) {
          try {
            const assignmentsResponse = await axios.get(`/api/sso/groups/${group.group_id}/assignments`)
            group.account_assignments = assignmentsResponse.data
          } catch (err) {
            group.account_assignments = []
          }
        }
        this.ssoGroups = groups
      } catch (err) {
        if (err.response?.status === 404 || err.response?.status === 503) {
          const errorMsg = err.response?.data?.error || 'SSO service is not available'
          const details = err.response?.data?.details || 'Ensure IAM Identity Center is enabled and the service has the required permissions.'
          // If error message already contains details, don't duplicate
          if (errorMsg.includes('Identity Center') || errorMsg.includes('permissions')) {
            this.error = errorMsg
          } else {
            this.error = `${errorMsg}. ${details}`
          }
        } else {
          this.error = err.response?.data?.error || err.response?.data?.details || err.message || 'Failed to load SSO groups'
        }
      } finally {
        this.loading = false
      }
    },
    refreshData() {
      this.loadData()
    },
    downloadJSON() {
      const dataStr = JSON.stringify(this.ssoGroups, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `sso-groups-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    },
    viewGroup(groupId) {
      this.$router.push(`/aws/sso/groups/${groupId}`)
    }
  }
}
</script>

<style scoped>
.sso-groups-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 15px;
}

.header-icon {
  width: 40px;
  height: 40px;
  color: #3b82f6;
}

.title-content h1 {
  margin: 0;
  font-size: 24px;
}

.title-content p {
  margin: 5px 0 0 0;
  color: var(--color-text-secondary);
}

.search-box {
  margin-bottom: 20px;
  position: relative;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  color: #9ca3af;
}

.search-input {
  width: 100%;
  padding: 10px 10px 10px 40px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

.search-input::placeholder {
  color: var(--color-text-tertiary);
}

.groups-table-container {
  overflow-x: auto;
}

.groups-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-bg-primary);
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.groups-table thead {
  background: var(--color-bg-secondary);
}

.groups-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-primary);
  border-bottom: 2px solid var(--color-border);
}

.groups-table tbody tr {
  background: var(--color-bg-primary);
}

.groups-table td {
  padding: 12px;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-primary);
  background: var(--color-bg-primary);
}

.groups-table tbody tr:hover {
  background: var(--color-bg-secondary);
}

.groups-table tbody tr:hover td {
  background: var(--color-bg-secondary);
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

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
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

.empty-results {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}
</style>
