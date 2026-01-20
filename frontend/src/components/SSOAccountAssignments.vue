<template>
  <div class="sso-account-assignments-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <div class="header-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12,1L8,5H11V14H13V5H16M18,23L22,19H19V10H17V19H14M2,13V11H5V13H2M7,13V11H10V13H7M12,13V11H15V13H12M17,13V11H20V13H17M2,18V16H5V18H2M7,18V16H10V18H7M12,18V16H15V18H12M17,18V16H20V18H17Z"/>
            </svg>
          </div>
          <div class="title-content">
            <h1>Account Assignments</h1>
            <p>{{ accounts.length }} accounts</p>
          </div>
        </div>
        <div class="header-actions">
          <button @click="downloadJSON" class="btn btn-secondary" :disabled="loading || accounts.length === 0">
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
      <button @click="refreshData" class="btn btn-primary">Try Again</button>
    </div>

    <div v-else class="main-content">
      <div class="filters-box">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" fill="currentColor">
            <path d="M9.5,3A6.5,6.5 0 0,1 16,9.5C16,11.11 15.41,12.59 14.44,13.73L14.71,14H15.5L20.5,19L19,20.5L14,15.5V14.71L13.73,14.44C12.59,15.41 11.11,16 9.5,16A6.5,6.5 0 0,1 3,9.5A6.5,6.5 0 0,1 9.5,3M9.5,5C7,5 5,7 5,9.5C5,12 7,14 9.5,14C12,14 14,12 14,9.5C14,7 12,5 9.5,5Z"/>
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by account name or ID..."
            class="search-input"
          />
        </div>
        <div class="sort-controls">
          <label>Sort by:</label>
          <select v-model="sortBy" class="sort-select">
            <option value="account_name">Account Name</option>
            <option value="account_id">Account ID</option>
            <option value="assignments_count">Number of Assignments</option>
          </select>
          <button @click="sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'" class="btn btn-sm btn-secondary">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path v-if="sortOrder === 'asc'" d="M7.41,15.41L12,10.83L16.59,15.41L18,14L12,8L6,14L7.41,15.41Z"/>
              <path v-else d="M7.41,8.58L12,13.17L16.59,8.58L18,10L12,16L6,10L7.41,8.58Z"/>
            </svg>
            {{ sortOrder === 'asc' ? 'Asc' : 'Desc' }}
          </button>
        </div>
      </div>

      <div v-if="filteredAccounts.length === 0 && !loading && !error" class="empty-results">
        <h4>No accounts found</h4>
        <p v-if="accounts.length === 0">No accounts are available. Make sure AWS Organization is configured.</p>
        <p v-else>No accounts match your search criteria.</p>
      </div>

      <div v-else class="accounts-table-container">
        <table class="accounts-table">
          <thead>
            <tr>
              <th>Account Name</th>
              <th>Account ID</th>
              <th>Assignments</th>
              <th>Users/Groups</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="account in sortedAccounts" :key="account.account_id">
              <td>{{ account.account_name || '-' }}</td>
              <td>{{ account.account_id }}</td>
              <td>
                <span class="badge">{{ account.assignments?.length || 0 }} assignments</span>
              </td>
              <td>
                <div class="assignments-list">
                  <span 
                    v-for="assignment in (account.assignments || [])" 
                    :key="`${account.account_id}-${assignment.principal_id}-${assignment.permission_set_arn}`"
                    class="assignment-badge"
                    :class="{ 'user-badge': assignment.principal_type === 'USER', 'group-badge': assignment.principal_type === 'GROUP' }"
                  >
                    {{ assignment.principal_name || assignment.principal_id }}
                    <span class="permission-set">({{ assignment.permission_set_name || 'N/A' }})</span>
                  </span>
                  <span v-if="!account.assignments || account.assignments.length === 0" class="no-assignments">No assignments</span>
                </div>
              </td>
              <td>
                <button 
                  v-if="account.assignments && account.assignments.length > 0"
                  @click="viewAccountDetails(account.account_id)"
                  class="btn btn-sm btn-primary"
                >
                  View Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'SSOAccountAssignments',
  data() {
    return {
      accounts: [],
      loading: true,
      error: null,
      searchQuery: '',
      sortBy: 'account_name',
      sortOrder: 'asc'
    }
  },
  computed: {
    filteredAccounts() {
      if (!this.searchQuery) return this.accounts
      const query = this.searchQuery.toLowerCase()
      return this.accounts.filter(account => 
        account.account_name?.toLowerCase().includes(query) ||
        account.account_id?.toLowerCase().includes(query)
      )
    },
    sortedAccounts() {
      const filtered = [...this.filteredAccounts]
      filtered.sort((a, b) => {
        let aVal, bVal
        
        switch (this.sortBy) {
          case 'account_name':
            aVal = (a.account_name || '').toLowerCase()
            bVal = (b.account_name || '').toLowerCase()
            break
          case 'account_id':
            aVal = a.account_id.toLowerCase()
            bVal = b.account_id.toLowerCase()
            break
          case 'assignments_count':
            aVal = a.assignments?.length || 0
            bVal = b.assignments?.length || 0
            break
          default:
            return 0
        }
        
        if (aVal < bVal) return this.sortOrder === 'asc' ? -1 : 1
        if (aVal > bVal) return this.sortOrder === 'asc' ? 1 : -1
        return 0
      })
      return filtered
    }
  },
  mounted() {
    // Restore filter state from query params
    const query = this.$route.query
    if (query.search) {
      this.searchQuery = query.search
    }
    if (query.sortBy) {
      this.sortBy = query.sortBy
    }
    if (query.sortOrder) {
      this.sortOrder = query.sortOrder
    }
    this.loadData()
  },
  watch: {
    searchQuery(newVal) {
      this.updateQueryParams()
    },
    sortBy(newVal) {
      this.updateQueryParams()
    },
    sortOrder(newVal) {
      this.updateQueryParams()
    }
  },
  methods: {
    updateQueryParams() {
      const query = {}
      if (this.searchQuery) {
        query.search = this.searchQuery
      }
      if (this.sortBy !== 'account_name') {
        query.sortBy = this.sortBy
      }
      if (this.sortOrder !== 'asc') {
        query.sortOrder = this.sortOrder
      }
      this.$router.replace({ query })
    },
    async loadData() {
      this.loading = true
      this.error = null
      try {
        const response = await axios.get('/api/sso/account-assignments')
        this.accounts = Array.isArray(response.data) ? response.data.map(account => ({
          ...account,
          assignments: account.assignments || []
        })) : []
        console.log('Loaded accounts:', this.accounts.length, 'accounts')
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
    refreshData() {
      this.loadData()
    },
    downloadJSON() {
      const dataStr = JSON.stringify(this.accounts, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `sso-account-assignments-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    },
    viewAccountDetails(accountId) {
      // Preserve filter state in query params when navigating to details
      const query = {}
      if (this.searchQuery) {
        query.search = this.searchQuery
      }
      if (this.sortBy !== 'account_name') {
        query.sortBy = this.sortBy
      }
      if (this.sortOrder !== 'asc') {
        query.sortOrder = this.sortOrder
      }
      this.$router.push({
        path: `/aws/sso/accounts/${accountId}/assignments`,
        query
      })
    }
  }
}
</script>

<style scoped>
.sso-account-assignments-container {
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

.filters-box {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 250px;
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

.sort-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.sort-controls label {
  color: var(--color-text-secondary);
  font-size: 14px;
}

.sort-select {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  cursor: pointer;
}

.accounts-table-container {
  overflow-x: auto;
}

.accounts-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-bg-primary);
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.accounts-table thead {
  background: var(--color-bg-secondary);
}

.accounts-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-primary);
  border-bottom: 2px solid var(--color-border);
  cursor: pointer;
  user-select: none;
}

.accounts-table tbody tr {
  background: var(--color-bg-primary);
}

.accounts-table td {
  padding: 12px;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-primary);
  background: var(--color-bg-primary);
}

.accounts-table tbody tr:hover {
  background: var(--color-bg-secondary);
}

.accounts-table tbody tr:hover td {
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

.assignments-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.assignment-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: #3b82f6;
  color: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.assignment-badge.user-badge {
  background: #10b981;
}

.assignment-badge.group-badge {
  background: #8b5cf6;
}

.assignment-badge .permission-set {
  opacity: 0.9;
  font-size: 10px;
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

.btn-icon {
  width: 16px;
  height: 16px;
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

.empty-results {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}

.empty-results h4 {
  color: var(--color-text-primary);
  margin-bottom: 10px;
}
</style>
