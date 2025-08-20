<template>
  <div class="users-container">
    <header class="page-header">
      <div class="header-content">
        <div class="breadcrumb">
          <router-link to="/" class="breadcrumb-link">
            <svg class="breadcrumb-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-1 9H9V9h10v2zm-4 4H9v-2h6v2zm4-8H9V5h10v2z"/>
            </svg>
            Accounts
          </router-link>
          <svg class="breadcrumb-separator" viewBox="0 0 24 24" fill="currentColor">
            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/>
          </svg>
          <span class="breadcrumb-current">{{ accountId }}</span>
        </div>
        <div class="header-text">
          <div class="header-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2ZM21 9V7L15 1H5C3.89 1 3 1.89 3 3V18A2 2 0 0 0 5 20H11V18H5V3H13V9H21Z"/>
              <path d="M16 4c0-1.11.89-2 2-2s2 .89 2 2-.89 2-2 2-2-.89-2-2zM4 18v-4h3v4h2v-7.5c0-1.1.9-2 2-2s2 .9 2 2V11h2c0-.55-.45-1-1-1H8c-.55 0-1 .45-1 1v8H4z"/>
            </svg>
          </div>
          <div>
            <h2>IAM Users</h2>
            <p>Manage users and their access keys for account {{ accountId }}</p>
          </div>
        </div>
      </div>
      <div class="stats" v-if="!loading && !error">
        <div class="stat-item">
          <span class="stat-number">{{ users.length }}</span>
          <span class="stat-label">{{ users.length === 1 ? 'User' : 'Users' }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-number">{{ usersWithPasswords }}</span>
          <span class="stat-label">With Passwords</span>
        </div>
        <div class="stat-item">
          <span class="stat-number">{{ totalAccessKeys }}</span>
          <span class="stat-label">Access Keys</span>
        </div>
      </div>
    </header>

    <div v-if="loading" class="loading">
      <span>Loading users...</span>
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="users.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2ZM16 4c0-1.11.89-2 2-2s2 .89 2 2-.89 2-2 2-2-.89-2-2zM4 18v-4h3v4h2v-7.5c0-1.1.9-2 2-2s2 .9 2 2V11h2c0-.55-.45-1-1-1H8c-.55 0-1 .45-1 1v8H4z"/>
        </svg>
      </div>
      <h3>No users found</h3>
      <p>There are no IAM users in this AWS account.</p>
    </div>
    <div v-else class="table-container">
      <div class="table-header">
        <h3>Users Overview</h3>
        <div class="table-actions">
          <button class="btn btn-secondary" @click="refreshUsers">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
            </svg>
            Refresh
          </button>
        </div>
      </div>
      <div class="table-wrapper">
        <table class="modern-table">
          <thead>
            <tr>
              <th>
                <div class="th-content">
                  <svg class="th-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2Z"/>
                  </svg>
                  Username
                </div>
              </th>
              <th>
                <div class="th-content">
                  <svg class="th-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 1L3 5V11C3 16.55 6.84 21.74 12 23C17.16 21.74 21 16.55 21 11V5L12 1Z"/>
                  </svg>
                  User ID
                </div>
              </th>
              <th>
                <div class="th-content">
                  <svg class="th-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12,17A2,2 0 0,0 14,15C14,13.89 13.1,13 12,13A2,2 0 0,0 10,15A2,2 0 0,0 12,17M18,8A2,2 0 0,1 20,10V20A2,2 0 0,1 18,22H6A2,2 0 0,1 4,20V10C4,8.89 4.9,8 6,8H7V6A5,5 0 0,1 12,1A5,5 0 0,1 17,6V8H18M12,3A3,3 0 0,0 9,6V8H15V6A3,3 0 0,0 12,3Z"/>
                  </svg>
                  Password
                </div>
              </th>
              <th>
                <div class="th-content">
                  <svg class="th-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M7 14C5.9 14 5 13.1 5 12S5.9 10 7 10 9 10.9 9 12 8.1 14 7 14M12.6 10C11.8 7.7 9.6 6 7 6C3.7 6 1 8.7 1 12S3.7 18 7 18C9.6 18 11.8 16.3 12.6 14H16L17.5 15.5L19 14L17.5 12.5L19 11L17.5 9.5L16 11H12.6Z"/>
                  </svg>
                  Access Keys
                </div>
              </th>
              <th>
                <div class="th-content">
                  <svg class="th-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M19 3H18V1H16V3H8V1H6V3H5C3.89 3 3.01 3.9 3.01 5L3 19C3 20.1 3.89 21 5 21H19C20.1 21 21 20.1 21 19V5C21 3.9 20.1 3 19 3ZM19 19H5V8H19V19ZM7 10H12V15H7V10Z"/>
                  </svg>
                  Created
                </div>
              </th>
              <th>
                <div class="th-content">
                  Actions
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.username" class="table-row">
              <td>
                <div class="user-cell">
                  <div class="user-avatar">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2ZM21 9V7L15 1H5C3.89 1 3 1.89 3 3V18A2 2 0 0 0 5 20H11V18H5V3H13V9H21Z"/>
                    </svg>
                  </div>
                  <div class="user-info">
                    <span class="user-name">{{ user.username }}</span>
                  </div>
                </div>
              </td>
              <td>
                <code class="user-id">{{ user.user_id }}</code>
              </td>
              <td>
                <div class="status-cell">
                  <span :class="['status-badge', user.password_set ? 'status-success' : 'status-warning']">
                    <svg class="status-icon" viewBox="0 0 24 24" fill="currentColor">
                      <path v-if="user.password_set" d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"/>
                      <path v-else d="M19,13H13V11H19V13Z"/>
                    </svg>
                    {{ user.password_set ? 'Set' : 'None' }}
                  </span>
                </div>
              </td>
              <td>
                <div class="access-keys-cell">
                  <span class="access-keys-count">{{ user.access_keys.length }}</span>
                  <span class="access-keys-label">{{ user.access_keys.length === 1 ? 'key' : 'keys' }}</span>
                </div>
              </td>
              <td>
                <div class="date-cell">
                  <span class="date-value">{{ formatDate(user.create_date) }}</span>
                  <span class="date-relative">{{ formatRelativeDate(user.create_date) }}</span>
                </div>
              </td>
              <td>
                <div class="actions-cell">
                  <router-link :to="`/accounts/${accountId}/users/${user.username}`" class="btn btn-primary">
                    <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 4.5C7 4.5 2.73 7.61 1 12C2.73 16.39 7 19.5 12 19.5S21.27 16.39 23 12C21.27 7.61 17 4.5 12 4.5ZM12 17C9.24 17 7 14.76 7 12S9.24 7 12 7S17 9.24 17 12S14.76 17 12 17ZM12 9C10.34 9 9 10.34 9 12S10.34 15 12 15S15 13.66 15 12S13.66 9 12 9Z"/>
                    </svg>
                    View Details
                  </router-link>
                </div>
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
  name: 'Users',
  props: ['accountId'],
  data() {
    return {
      users: [],
      loading: true,
      error: null
    }
  },
  computed: {
    usersWithPasswords() {
      return this.users.filter(user => user.password_set).length
    },
    totalAccessKeys() {
      return this.users.reduce((total, user) => total + user.access_keys.length, 0)
    }
  },
  async mounted() {
    await this.loadUsers()
  },
  methods: {
    async loadUsers() {
      try {
        this.loading = true
        this.error = null
        const response = await axios.get(`/api/accounts/${this.accountId}/users`)
        this.users = response.data
      } catch (err) {
        this.error = err.response?.data?.error || 'Failed to load users'
      } finally {
        this.loading = false
      }
    },
    async refreshUsers() {
      await this.loadUsers()
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
      
      if (diffDays === 0) return 'Today'
      if (diffDays === 1) return 'Yesterday'
      if (diffDays < 30) return `${diffDays} days ago`
      if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`
      return `${Math.floor(diffDays / 365)} years ago`
    }
  }
}
</script>

<style scoped>
.users-container {
  animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-xl);
  padding: var(--spacing-lg);
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: 0 2px 8px var(--color-shadow-light);
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

/* Breadcrumb */
.breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--color-btn-primary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.breadcrumb-link:hover {
  color: var(--color-btn-primary-hover);
}

.breadcrumb-icon {
  width: 1rem;
  height: 1rem;
}

.breadcrumb-separator {
  width: 1rem;
  height: 1rem;
  color: var(--color-text-tertiary);
}

.breadcrumb-current {
  font-weight: 500;
  color: var(--color-text-primary);
}

/* Header Text */
.header-text {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header-icon {
  width: 3rem;
  height: 3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-accent);
  border-radius: var(--radius-lg);
  color: var(--color-text-inverse);
}

.header-icon svg {
  width: 1.5rem;
  height: 1.5rem;
}

.header-text h2 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.header-text p {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  margin: var(--spacing-xs) 0 0 0;
}

/* Stats */
.stats {
  display: flex;
  gap: var(--spacing-lg);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-md);
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  min-width: 80px;
}

.stat-number {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-btn-primary);
  line-height: 1;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-top: var(--spacing-xs);
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: var(--spacing-xxl);
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  margin-top: var(--spacing-xl);
}

.empty-icon {
  width: 4rem;
  height: 4rem;
  margin: 0 auto var(--spacing-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-secondary);
  border-radius: 50%;
  color: var(--color-text-secondary);
}

.empty-icon svg {
  width: 2rem;
  height: 2rem;
}

.empty-state h3 {
  font-size: 1.25rem;
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-sm);
}

.empty-state p {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

/* Table Container */
.table-container {
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 12px var(--color-shadow-light);
  border: 1px solid var(--color-border-light);
  overflow: hidden;
  margin-top: var(--spacing-lg);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  background: var(--color-bg-secondary);
}

.table-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.table-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.table-wrapper {
  overflow-x: auto;
}

/* Modern Table */
.modern-table {
  width: 100%;
  border-collapse: collapse;
}

.modern-table th {
  background: var(--color-bg-secondary);
  padding: var(--spacing-md) var(--spacing-lg);
  text-align: left;
  border-bottom: 2px solid var(--color-border-light);
  position: sticky;
  top: 0;
  z-index: 10;
}

.th-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.th-icon {
  width: 1rem;
  height: 1rem;
  color: var(--color-text-secondary);
}

.modern-table td {
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  vertical-align: middle;
}

.table-row {
  transition: background-color var(--transition-fast);
}

.table-row:hover {
  background-color: var(--color-bg-secondary);
}

/* Cell Styles */
.user-cell {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.user-avatar {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-btn-primary), var(--color-bg-accent));
  border-radius: var(--radius-md);
  color: var(--color-text-inverse);
  flex-shrink: 0;
}

.user-avatar svg {
  width: 1.25rem;
  height: 1.25rem;
}

.user-info {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.9rem;
}

.user-id {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.75rem;
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border-light);
}

.status-cell {
  display: flex;
  align-items: center;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-badge.status-success {
  background: rgba(40, 167, 69, 0.1);
  color: var(--color-success);
  border: 1px solid rgba(40, 167, 69, 0.2);
}

.status-badge.status-warning {
  background: rgba(255, 193, 7, 0.1);
  color: var(--color-warning);
  border: 1px solid rgba(255, 193, 7, 0.2);
}

.status-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.access-keys-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.access-keys-count {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  line-height: 1;
}

.access-keys-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-top: 2px;
}

.date-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.date-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
  line-height: 1;
}

.date-relative {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.actions-cell {
  display: flex;
  gap: var(--spacing-xs);
}

.btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.btn-icon {
  width: 1rem;
  height: 1rem;
}

.btn-primary {
  background: var(--color-btn-primary);
  color: var(--color-text-inverse);
}

.btn-primary:hover {
  background: var(--color-btn-primary-hover);
}

.btn-secondary {
  background: var(--color-btn-secondary);
  color: var(--color-text-inverse);
}

.btn-secondary:hover {
  background: var(--color-btn-secondary-hover);
}

/* Responsive Design */
@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .stats {
    justify-content: center;
  }
}

@media (max-width: 768px) {
  .header-text {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .table-header {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .table-actions {
    justify-content: center;
  }
  
  .stats {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-md);
  }
  
  .modern-table {
    font-size: 0.875rem;
  }
  
  .modern-table th,
  .modern-table td {
    padding: var(--spacing-sm) var(--spacing-md);
  }
  
  .user-cell {
    gap: var(--spacing-sm);
  }
  
  .user-avatar {
    width: 2rem;
    height: 2rem;
  }
  
  .user-avatar svg {
    width: 1rem;
    height: 1rem;
  }
}

@media (max-width: 480px) {
  .breadcrumb {
    font-size: 0.75rem;
  }
  
  .header-text h2 {
    font-size: 1.5rem;
  }
  
  .stats {
    grid-template-columns: 1fr;
  }
  
  .stat-number {
    font-size: 1.5rem;
  }
  
  /* Hide some table columns on very small screens */
  .modern-table th:nth-child(2),
  .modern-table td:nth-child(2),
  .modern-table th:nth-child(5),
  .modern-table td:nth-child(5) {
    display: none;
  }
}

/* Loading and Error States */
.loading {
  text-align: center;
  padding: var(--spacing-xxl);
  color: var(--color-text-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  margin-top: var(--spacing-lg);
}

.loading::before {
  content: '';
  width: 2rem;
  height: 2rem;
  border: 2px solid var(--color-border);
  border-top: 2px solid var(--color-btn-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error {
  background: rgba(220, 53, 69, 0.1);
  color: var(--color-danger);
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  margin: var(--spacing-lg) 0;
  border: 1px solid rgba(220, 53, 69, 0.2);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.error::before {
  content: '⚠';
  font-size: 1.2rem;
}
</style>