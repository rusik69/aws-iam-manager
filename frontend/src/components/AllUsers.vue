<template>
  <div class="all-users-container">
    <!-- Header -->
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
            <h1>All Organization Users</h1>
            <p>{{ allUsers.length }} users across {{ accounts.length }} accounts</p>
          </div>
        </div>
        <div class="header-actions">
          <button @click="refreshData" class="btn btn-secondary" :disabled="loading">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
            </svg>
            {{ loading ? 'Refreshing...' : 'Refresh' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Loading organization users...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-container">
      <div class="error-icon">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
        </svg>
      </div>
      <h3>Failed to Load Users</h3>
      <p>{{ error }}</p>
      <button @click="refreshData" class="btn btn-primary">Try Again</button>
    </div>

    <!-- Main Content -->
    <div v-else class="main-content">
      <!-- Search and Filter -->
      <div class="user-filters">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" fill="currentColor">
            <path d="M9.5,3A6.5,6.5 0 0,1 16,9.5C16,11.11 15.41,12.59 14.44,13.73L14.71,14H15.5L20.5,19L19,20.5L14,15.5V14.71L13.73,14.44C12.59,15.41 11.11,16 9.5,16A6.5,6.5 0 0,1 3,9.5A6.5,6.5 0 0,1 9.5,3M9.5,5C7,5 5,7 5,9.5C5,12 7,14 9.5,14C12,14 14,12 14,9.5C14,7 12,5 9.5,5Z"/>
          </svg>
          <input 
            v-model="userSearchQuery" 
            type="text" 
            placeholder="Search users..."
            class="search-input"
          />
        </div>
        <div class="filter-buttons">
          <button 
            @click="userFilter = 'all'" 
            :class="['filter-btn', { active: userFilter === 'all' }]"
          >
            All ({{ allUsers.length }})
          </button>
          <button 
            @click="userFilter = 'withKeys'" 
            :class="['filter-btn', { active: userFilter === 'withKeys' }]"
          >
            With Keys ({{ usersWithKeys }})
          </button>
          <button 
            @click="userFilter = 'oldKeys'" 
            :class="['filter-btn', { active: userFilter === 'oldKeys' }]"
          >
            Old Keys ({{ usersWithOldKeys }})
          </button>
          <button 
            @click="userFilter = 'withPasswords'" 
            :class="['filter-btn', { active: userFilter === 'withPasswords' }]"
          >
            With Passwords ({{ usersWithPasswords }})
          </button>
        </div>
      </div>

      <!-- Users Table -->
      <div class="users-table-container">
        <table class="users-table">
          <thead>
            <tr>
              <th>User</th>
              <th>Account</th>
              <th>Password</th>
              <th>Keys</th>
              <th>Old Keys</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in filteredUsers" :key="`${user.accountId}-${user.username}`" class="user-row">
              <td>
                <div class="user-info">
                  <div class="user-name">{{ user.username }}</div>
                  <div class="user-id">{{ user.user_id }}</div>
                </div>
              </td>
              <td>
                <div class="account-info">
                  <div class="account-name">{{ user.accountName }}</div>
                  <div class="account-id">{{ user.accountId }}</div>
                </div>
              </td>
              <td>
                <span :class="['password-status', user.password_set ? 'has-password' : 'no-password']">
                  <svg class="status-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path v-if="user.password_set" d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"/>
                    <path v-else d="M19,13H5V11H19V13Z"/>
                  </svg>
                  {{ user.password_set ? 'Yes' : 'No' }}
                </span>
              </td>
              <td>
                <span class="key-count">{{ user.access_keys?.length || 0 }}</span>
              </td>
              <td>
                <span :class="['old-key-count', { warning: getOldKeyCount(user) > 0 }]">
                  {{ getOldKeyCount(user) }}
                </span>
              </td>
              <td>
                <div class="date-info">
                  <div class="date-value">{{ formatDate(user.create_date) }}</div>
                  <div class="date-relative">{{ formatRelativeDate(user.create_date) }}</div>
                </div>
              </td>
              <td>
                <div class="user-actions">
                  <button 
                    @click="viewUser(user.accountId, user.username)"
                    class="btn btn-sm btn-primary"
                  >
                    <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12,9A3,3 0 0,0 9,12A3,3 0 0,0 12,15A3,3 0 0,0 15,12A3,3 0 0,0 12,9M12,17A5,5 0 0,1 7,12A5,5 0 0,1 12,7A5,5 0 0,1 17,12A5,5 0 0,1 12,17M12,4.5C7,4.5 2.73,7.61 1,12C2.73,16.39 7,19.5 12,19.5C17,19.5 21.27,16.39 23,12C21.27,7.61 17,4.5 12,4.5Z"/>
                    </svg>
                    View
                  </button>
                  <button 
                    @click="deleteUser(user.accountId, user.username)"
                    class="btn btn-sm btn-danger"
                    :disabled="actionLoading[`${user.accountId}-${user.username}`]"
                  >
                    <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M19,4H15.5L14.5,3H9.5L8.5,4H5V6H19M6,19A2,2 0 0,0 8,21H16A2,2 0 0,0 18,19V7H6V19Z"/>
                    </svg>
                    {{ actionLoading[`${user.accountId}-${user.username}`] ? 'Deleting...' : 'Delete' }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        
        <div v-if="filteredUsers.length === 0" class="empty-results">
          <div class="empty-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2Z"/>
            </svg>
          </div>
          <h4>No users found</h4>
          <p>No users match the current search and filter criteria.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'AllUsers',
  data() {
    return {
      accounts: [],
      allUsers: [],
      loading: true,
      error: null,
      userSearchQuery: '',
      userFilter: 'all',
      actionLoading: {}
    }
  },
  computed: {
    usersWithKeys() {
      return this.allUsers?.filter(user => (user.access_keys?.length || 0) > 0)?.length || 0
    },
    usersWithOldKeys() {
      const thirtyDaysAgo = new Date()
      thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)
      
      return this.allUsers?.filter(user => {
        return user.access_keys?.some(key => {
          const keyDate = new Date(key.create_date)
          return keyDate < thirtyDaysAgo
        })
      })?.length || 0
    },
    usersWithPasswords() {
      return this.allUsers?.filter(user => user.password_set)?.length || 0
    },
    enrichedUsers() {
      return this.allUsers?.map(user => {
        const account = this.accounts?.find(acc => acc.id === user.accountId)
        return {
          ...user,
          accountName: account?.name || 'Unknown Account'
        }
      }) || []
    },
    filteredUsers() {
      let users = this.enrichedUsers

      // Apply search filter
      if (this.userSearchQuery) {
        const query = this.userSearchQuery.toLowerCase()
        users = users.filter(user => 
          user.username.toLowerCase().includes(query) ||
          user.user_id.toLowerCase().includes(query) ||
          user.accountName.toLowerCase().includes(query) ||
          user.accountId.toLowerCase().includes(query)
        )
      }

      // Apply category filter
      const thirtyDaysAgo = new Date()
      thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

      switch (this.userFilter) {
        case 'withKeys':
          return users.filter(user => (user.access_keys?.length || 0) > 0)
        case 'oldKeys':
          return users.filter(user => 
            user.access_keys?.some(key => {
              const keyDate = new Date(key.create_date)
              return keyDate < thirtyDaysAgo
            })
          )
        case 'withPasswords':
          return users.filter(user => user.password_set)
        default:
          return users
      }
    }
  },
  async mounted() {
    await this.loadData()
  },
  methods: {
    async loadData() {
      try {
        this.loading = true
        this.error = null
        
        // Load accounts first
        const accountsResponse = await axios.get('/api/accounts')
        this.accounts = accountsResponse.data
        
        // Load users for all accounts
        await this.loadAllUsers()
        
      } catch (err) {
        this.error = err.response?.data?.error || 'Failed to load data'
      } finally {
        this.loading = false
      }
    },
    
    async loadAllUsers() {
      if (!this.accounts || this.accounts.length === 0) return
      
      try {
        const allUsers = []
        
        // Load users for each account in parallel
        const userPromises = this.accounts.map(async (account) => {
          try {
            const response = await axios.get(`/api/accounts/${account.id}/users`)
            const users = response.data || []
            
            // Add account info to users
            const usersWithAccountInfo = users.map(user => ({
              ...user,
              accountId: account.id,
              accountName: account.name
            }))
            allUsers.push(...usersWithAccountInfo)
            
          } catch (err) {
            console.warn(`Failed to load users for account ${account.id}:`, err)
          }
        })
        
        await Promise.all(userPromises)
        this.allUsers = allUsers
        
      } catch (err) {
        console.error('Failed to load users:', err)
      }
    },
    
    async refreshData() {
      this.allUsers = []
      await this.loadData()
    },
    
    getOldKeyCount(user) {
      if (!user.access_keys || user.access_keys.length === 0) return 0
      
      const thirtyDaysAgo = new Date()
      thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)
      
      return user.access_keys.filter(key => {
        const keyDate = new Date(key.create_date)
        return keyDate < thirtyDaysAgo
      }).length
    },
    
    formatDate(dateString) {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      })
    },
    
    formatRelativeDate(dateString) {
      const now = new Date()
      const date = new Date(dateString)
      const diffTime = Math.abs(now - date)
      const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
      
      if (diffDays === 0) return 'today'
      if (diffDays === 1) return 'yesterday'
      if (diffDays < 30) return `${diffDays} days ago`
      if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`
      return `${Math.floor(diffDays / 365)} years ago`
    },
    
    viewUser(accountId, username) {
      this.$router.push(`/accounts/${accountId}/users/${username}`)
    },
    
    async deleteUser(accountId, username) {
      const confirmMessage = `Are you sure you want to DELETE the user "${username}"?\n\nThis will:\n1. Delete all access keys for this user\n2. Delete the user's login profile (if exists)\n3. Permanently delete the user\n\nThis action cannot be undone!`
      
      if (!confirm(confirmMessage)) return
      
      // Second confirmation for safety
      const finalConfirmation = confirm(`FINAL CONFIRMATION:\nType "DELETE" to confirm deletion of user "${username}"`)
      if (!finalConfirmation) return
      
      const loadingKey = `${accountId}-${username}`
      this.actionLoading[loadingKey] = true
      
      try {
        await axios.delete(`/api/accounts/${accountId}/users/${username}`)
        alert(`User "${username}" has been successfully deleted.`)
        // Refresh the data to update the list
        await this.refreshData()
      } catch (error) {
        console.error('Failed to delete user:', error)
        alert(error.response?.data?.error || 'Failed to delete user. Please try again.')
      } finally {
        delete this.actionLoading[loadingKey]
      }
    }
  }
}
</script>

<style scoped>
/* Container */
.all-users-container {
  min-height: 100vh;
  background: var(--color-bg-primary, #ffffff);
  padding: 1.5rem;
}

/* Page Header */
.page-header {
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: var(--color-bg-secondary, #f9fafb);
  border-radius: 0.75rem;
  border: 1px solid var(--color-border-light, #e5e7eb);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-icon {
  width: 4rem;
  height: 4rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  border-radius: 1rem;
  color: white;
  flex-shrink: 0;
}

.header-icon svg {
  width: 2rem;
  height: 2rem;
}

.title-content h1 {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text-primary, #1f2937);
  margin: 0 0 0.5rem 0;
}

.title-content p {
  font-size: 1rem;
  color: var(--color-text-secondary, #6b7280);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
  flex-shrink: 0;
}

/* Loading & Error States */
.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  text-align: center;
  padding: 3rem;
}

.loading-spinner {
  width: 3rem;
  height: 3rem;
  border: 3px solid var(--color-border-light, #e5e7eb);
  border-top: 3px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-icon {
  width: 4rem;
  height: 4rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 50%;
  color: #ef4444;
  margin-bottom: 1rem;
}

.error-icon svg {
  width: 2rem;
  height: 2rem;
}

.error-container h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary, #1f2937);
  margin-bottom: 0.5rem;
}

.error-container p {
  color: var(--color-text-secondary, #6b7280);
  margin-bottom: 1.5rem;
}

/* Main Content */
.main-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* User Filters */
.user-filters {
  padding: 1.5rem;
  background: var(--color-bg-secondary, #f9fafb);
  border-radius: 0.75rem;
  border: 1px solid var(--color-border-light, #e5e7eb);
}

.search-box {
  position: relative;
  margin-bottom: 1rem;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  width: 1rem;
  height: 1rem;
  color: var(--color-text-secondary, #6b7280);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.75rem 0.75rem 0.75rem 2.5rem;
  border: 1px solid var(--color-border-light, #e5e7eb);
  border-radius: 0.5rem;
  font-size: 0.875rem;
  background: var(--color-bg-primary, #ffffff);
  color: var(--color-text-primary, #1f2937);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.filter-btn {
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-border-light, #e5e7eb);
  border-radius: 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  background: var(--color-bg-primary, #ffffff);
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.2s ease;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.filter-btn:hover {
  background: var(--color-bg-tertiary, #f3f4f6);
}

.filter-btn.active {
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: white;
  border-color: transparent;
}

/* Users Table */
.users-table-container {
  background: var(--color-bg-primary, #ffffff);
  border-radius: 0.75rem;
  border: 1px solid var(--color-border-light, #e5e7eb);
  overflow: hidden;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.users-table th {
  background: var(--color-bg-secondary, #f9fafb);
  color: var(--color-text-secondary, #6b7280);
  font-weight: 600;
  text-align: left;
  padding: 1rem 0.75rem;
  border-bottom: 1px solid var(--color-border-light, #e5e7eb);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.users-table td {
  padding: 1rem 0.75rem;
  border-bottom: 1px solid var(--color-border-light, #e5e7eb);
  vertical-align: top;
}

.user-row:hover {
  background: var(--color-bg-secondary, #f9fafb);
}

.user-info,
.account-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.user-name,
.account-name {
  font-weight: 500;
  color: var(--color-text-primary, #1f2937);
}

.user-id,
.account-id {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.75rem;
  color: var(--color-text-secondary, #6b7280);
  background: var(--color-bg-secondary, #f9fafb);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  border: 1px solid var(--color-border-light, #e5e7eb);
  width: fit-content;
}

.password-status {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.password-status.has-password {
  color: #059669;
}

.password-status.no-password {
  color: #dc2626;
}

.status-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.key-count {
  font-weight: 600;
  color: var(--color-text-primary, #1f2937);
}

.old-key-count {
  font-weight: 600;
  color: var(--color-text-primary, #1f2937);
}

.old-key-count.warning {
  color: #f59e0b;
}

.date-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.date-value {
  font-weight: 500;
  color: var(--color-text-primary, #1f2937);
}

.date-relative {
  font-size: 0.75rem;
  color: var(--color-text-secondary, #6b7280);
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
  justify-content: center;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb, #4f46e5);
  transform: translateY(-1px);
}

.btn-secondary {
  background: var(--color-bg-secondary, #f9fafb);
  color: var(--color-text-secondary, #6b7280);
  border: 1px solid var(--color-border-light, #e5e7eb);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--color-bg-tertiary, #f3f4f6);
  color: var(--color-text-primary, #1f2937);
}

.btn-sm {
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
  gap: 0.375rem;
}

.btn-icon {
  width: 1rem;
  height: 1rem;
}

/* Empty Results */
.empty-results {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--color-text-secondary, #6b7280);
}

.empty-icon {
  width: 4rem;
  height: 4rem;
  margin: 0 auto 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-secondary, #f9fafb);
  border-radius: 50%;
  color: var(--color-text-secondary, #6b7280);
}

.empty-icon svg {
  width: 2rem;
  height: 2rem;
}

.empty-results h4 {
  font-size: 1.125rem;
  color: var(--color-text-primary, #1f2937);
  margin-bottom: 0.5rem;
}

.empty-results p {
  color: var(--color-text-secondary, #6b7280);
}

/* Responsive Design */
@media (max-width: 1024px) {
  .users-table th:nth-child(3),
  .users-table td:nth-child(3),
  .users-table th:nth-child(6),
  .users-table td:nth-child(6) {
    display: none;
  }
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .header-actions {
    justify-content: center;
  }
  
  .filter-buttons {
    flex-direction: column;
  }
  
  .filter-btn {
    text-align: center;
  }
  
  .users-table th:nth-child(2),
  .users-table td:nth-child(2) {
    display: none;
  }
}

@media (max-width: 480px) {
  .all-users-container {
    padding: 1rem;
  }
  
  .header-title {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .header-icon {
    width: 3rem;
    height: 3rem;
  }
  
  .header-icon svg {
    width: 1.5rem;
    height: 1.5rem;
  }
  
  .title-content h1 {
    font-size: 1.5rem;
  }
  
  .users-table th:nth-child(4),
  .users-table td:nth-child(4),
  .users-table th:nth-child(5),
  .users-table td:nth-child(5) {
    display: none;
  }
}

/* Dark mode support */
.dark .all-users-container {
  background: #0d1117;
}

.dark .page-header,
.dark .user-filters,
.dark .users-table-container {
  background: #21262d;
  border-color: #30363d;
}

.dark .search-input {
  background: #21262d;
  border-color: #30363d;
  color: #ffffff;
}

.dark .filter-btn {
  background: #21262d;
  border-color: #30363d;
  color: #ffffff;
}

.dark .filter-btn:hover {
  background: #30363d;
}

.dark .users-table th {
  background: #161b22;
  border-color: #30363d;
  color: #ffffff;
}

.dark .users-table td {
  border-color: #30363d;
}

.dark .user-row:hover {
  background: #161b22;
}

.dark .user-id,
.dark .account-id {
  background: #161b22;
  border-color: #30363d;
  color: #ffffff;
}

.dark .btn-secondary {
  background: #21262d;
  color: #ffffff;
  border-color: #30363d;
}

.dark .btn-secondary:hover:not(:disabled) {
  background: #30363d;
}

.dark .empty-results,
.dark .empty-icon {
  background: #161b22;
}

.dark .error-container {
  background: #0d1117;
}

.user-actions {
  display: flex;
  gap: var(--spacing-xs);
  align-items: center;
}
</style>