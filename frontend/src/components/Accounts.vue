<template>
  <div>
    <div class="card">
      <h2>AWS Accounts</h2>
      <p>Found {{ accounts.length }} account(s)</p>
    </div>

    <div v-if="loading" class="loading">
      Loading accounts...
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="accounts.length === 0" class="card">
      <h3>No accounts found</h3>
      <p>There are no AWS accounts available in your organization.</p>
    </div>
    <div v-else class="accounts-grid">
      <div v-for="account in accounts" :key="account.id" class="card">
        <h3>{{ account.name }}</h3>
        <p><strong>ID:</strong> {{ account.id }}</p>
        <p><strong>Status:</strong> Active</p>
        <router-link :to="`/accounts/${account.id}/users`" class="btn">
          Manage Users
        </router-link>
        <button class="btn btn-secondary" @click="viewAccountDetails(account)">
          Details
        </button>
      </div>
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
      // TODO: Implement account details modal
    }
  }
}
</script>

<style scoped>
.accounts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

@media (max-width: 768px) {
  .accounts-grid {
    grid-template-columns: 1fr;
  }
}
</style>