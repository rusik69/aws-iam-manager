<template>
  <div class="accounts-container">
    <header class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-1 9H9V9h10v2zm-4 4H9v-2h6v2zm4-8H9V5h10v2z"/>
          </svg>
        </div>
        <div class="header-text">
          <h2>AWS Accounts</h2>
          <p>Manage IAM users across your AWS organization accounts</p>
        </div>
      </div>
      <div class="stats" v-if="!loading && !error">
        <div class="stat-item">
          <span class="stat-number">{{ accounts.length }}</span>
          <span class="stat-label">{{ accounts.length === 1 ? 'Account' : 'Accounts' }}</span>
        </div>
      </div>
    </header>

    <div v-if="loading" class="loading">
      <span>Loading accounts...</span>
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else class="accounts-grid">
      <div v-for="account in accounts" :key="account.id" class="account-card">
        <div class="account-card-header">
          <div class="account-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2ZM21 9V7L15 1H5C3.89 1 3 1.89 3 3V18A2 2 0 0 0 5 20H11V18H5V3H13V9H21Z"/>
            </svg>
          </div>
          <div class="account-status">
            <span class="status-badge status-active">Active</span>
          </div>
        </div>
        <div class="account-card-body">
          <h3 class="account-name">{{ account.name }}</h3>
          <div class="account-details">
            <div class="detail-row">
              <span class="detail-label">Account ID</span>
              <span class="detail-value">{{ account.id }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Region</span>
              <span class="detail-value">Global</span>
            </div>
          </div>
        </div>
        <div class="account-card-footer">
          <router-link :to="`/accounts/${account.id}/users`" class="btn btn-primary">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M16 4c0-1.11.89-2 2-2s2 .89 2 2-.89 2-2 2-2-.89-2-2zM4 18v-4h3v4h2v-7.5c0-1.1.9-2 2-2s2 .9 2 2V11h2c0-.55-.45-1-1-1H8c-.55 0-1 .45-1 1v8H4z"/>
            </svg>
            Manage Users
          </router-link>
          <button class="btn btn-secondary" @click="viewAccountDetails(account)">
            <svg class="btn-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
            Details
          </button>
        </div>
      </div>
    </div>
    
    <div v-if="!loading && accounts.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
        </svg>
      </div>
      <h3>No accounts found</h3>
      <p>There are no AWS accounts available in your organization.</p>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Accounts',
  data() {
    return {
      accounts: [],
      loading: true,
      error: null
    }
  },
  async mounted() {
    try {
      const response = await axios.get('/api/accounts')
      this.accounts = response.data
    } catch (err) {
      this.error = err.response?.data?.error || 'Failed to load accounts'
    } finally {
      this.loading = false
    }
  },
  methods: {
    viewAccountDetails(account) {
      // Future implementation for account details modal/view
      console.log('View account details:', account)
    }
  }
}
</script>

<style scoped>
.accounts-container {
  animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

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

.accounts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-lg);
  margin-top: var(--spacing-lg);
}

.account-card {
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 12px var(--color-shadow-light);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-normal);
  overflow: hidden;
  position: relative;
}

.account-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px var(--color-shadow);
}

.account-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg) var(--spacing-lg) var(--spacing-md);
}

.account-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-btn-primary), var(--color-bg-accent));
  border-radius: var(--radius-md);
  color: var(--color-text-inverse);
}

.account-icon svg {
  width: 1.25rem;
  height: 1.25rem;
}

.status-badge {
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-badge.status-active {
  background: rgba(40, 167, 69, 0.1);
  color: var(--color-success);
  border: 1px solid rgba(40, 167, 69, 0.2);
}

.account-card-body {
  padding: 0 var(--spacing-lg) var(--spacing-md);
}

.account-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-md);
}

.account-details {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: var(--color-bg-secondary);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
}

.account-card-footer {
  padding: var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
  background: var(--color-bg-secondary);
  display: flex;
  gap: var(--spacing-sm);
}

.btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  text-decoration: none;
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

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .accounts-grid {
    grid-template-columns: 1fr;
  }
  
  .account-card-footer {
    flex-direction: column;
  }
  
  .btn {
    flex: none;
  }
}
</style>