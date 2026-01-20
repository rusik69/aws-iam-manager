<template>
  <div class="sso-user-assignments-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <div class="header-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2Z"/>
              <path d="M21 9V7L15 1H5C3.89 1 3 1.89 3 3V19C3 20.11 3.89 21 5 21H11V19H5V3H13V9H21Z"/>
            </svg>
          </div>
          <div class="title-content">
            <h1>SSO User Assignments</h1>
            <p>View which users have access to which accounts</p>
          </div>
        </div>
        <div class="header-actions">
          <button @click="downloadJSON" class="btn btn-secondary" :disabled="loading || users.length === 0">
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
      <p>Loading user assignments...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <div class="error-icon">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
        </svg>
      </div>
      <h3>Failed to Load User Assignments</h3>
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
          placeholder="Search users or accounts..."
          class="search-input"
        />
      </div>

      <div v-if="users.length === 0 && !loading && !error" class="empty-results">
        <h4>No users found</h4>
        <p>No SSO users are available. Make sure IAM Identity Center has users configured.</p>
      </div>

      <div v-else class="users-table-container">
        <table class="users-table">
          <thead>
            <tr>
              <th>User</th>
              <th>Display Name</th>
              <th>Email</th>
              <th>Account Assignments</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in filteredUsers" :key="user.user_id">
              <td>{{ user.user_name }}</td>
              <td>{{ user.display_name || '-' }}</td>
              <td>{{ user.emails?.[0] || '-' }}</td>
              <td>
                <div class="assignments-list">
                  <span 
                    v-for="assignment in (user.account_assignments || [])" 
                    :key="`${assignment.account_id}-${assignment.permission_set_arn}`"
                    class="assignment-badge"
                  >
                    {{ assignment.account_name || assignment.account_id }} ({{ assignment.permission_set_name || 'N/A' }})
                  </span>
                  <span v-if="!user.account_assignments || user.account_assignments.length === 0" class="no-assignments">No assignments</span>
                </div>
              </td>
              <td>
                <button 
                  @click="viewUser(user.user_id)"
                  class="btn btn-sm btn-primary"
                >
                  View Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="filteredUsers.length === 0 && users.length > 0" class="empty-results">
          <h4>No users match your search criteria</h4>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'SSOUserAssignments',
  data() {
    return {
      users: [],
      loading: true,
      error: null,
      searchQuery: ''
    }
  },
  computed: {
    filteredUsers() {
      if (!this.searchQuery) return this.users
      const query = this.searchQuery.toLowerCase()
      return this.users.filter(user => {
        const matchesUser = user.user_name?.toLowerCase().includes(query) ||
          user.display_name?.toLowerCase().includes(query) ||
          user.emails?.some(email => email.toLowerCase().includes(query))
        const matchesAccount = user.account_assignments?.some(a => 
          a.account_id?.toLowerCase().includes(query) ||
          a.account_name?.toLowerCase().includes(query)
        )
        return matchesUser || matchesAccount
      })
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
        const response = await axios.get('/api/sso/user-assignments')
        // Ensure account_assignments is always an array
        this.users = Array.isArray(response.data) ? response.data.map(user => ({
          ...user,
          account_assignments: user.account_assignments || []
        })) : []
        console.log('Loaded users:', this.users.length, 'users')
      } catch (err) {
        console.error('Error loading user assignments:', err)
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
          this.error = err.response?.data?.error || err.response?.data?.details || err.message || 'Failed to load user assignments'
        }
      } finally {
        this.loading = false
      }
    },
    refreshData() {
      this.loadData()
    },
    downloadJSON() {
      const dataStr = JSON.stringify(this.users, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `sso-user-assignments-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    },
    viewUser(userId) {
      this.$router.push(`/aws/sso/users/${userId}`)
    }
  }
}
</script>

<style scoped>
.sso-user-assignments-container {
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
  color: var(--color-text-primary);
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
  color: var(--color-text-tertiary);
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

.users-table-container {
  overflow-x: auto;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-bg-primary);
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.users-table thead {
  background: var(--color-bg-secondary);
}

.users-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-primary);
  border-bottom: 2px solid var(--color-border);
}

.users-table tbody tr {
  background: var(--color-bg-primary);
}

.users-table td {
  padding: 12px;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-primary);
  background: var(--color-bg-primary);
}

.users-table tbody tr:hover {
  background: var(--color-bg-secondary);
}

.users-table tbody tr:hover td {
  background: var(--color-bg-secondary);
}

.assignments-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.assignment-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #3b82f6;
  color: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.no-assignments {
  color: var(--color-text-tertiary);
  font-size: 12px;
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

.empty-results {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}
</style>
