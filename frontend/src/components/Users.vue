<template>
  <div>
    <div class="card">
      <nav>
        <router-link to="/">← Accounts</router-link> / {{ accountId }}
      </nav>
      <h2>IAM Users</h2>
      <p>{{ users.length }} users, {{ usersWithPasswords }} with passwords, {{ totalAccessKeys }} access keys</p>
      <button class="btn btn-secondary" @click="refreshUsers">Refresh</button>
    </div>

    <div v-if="loading" class="loading">
      Loading users...
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="users.length === 0" class="card">
      <h3>No users found</h3>
      <p>There are no IAM users in this AWS account.</p>
    </div>
    <table v-else class="table">
      <thead>
        <tr>
          <th>Username</th>
          <th>User ID</th>
          <th>Password</th>
          <th>Keys</th>
          <th>Created</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.username">
          <td><strong>{{ user.username }}</strong></td>
          <td><code>{{ user.user_id }}</code></td>
          <td>
            <span :class="user.password_set ? 'status-success' : 'status-warning'">
              {{ user.password_set ? '✓ Set' : '✗ None' }}
            </span>
          </td>
          <td>{{ user.access_keys.length }}</td>
          <td>{{ formatDate(user.create_date) }}</td>
          <td>
            <router-link :to="`/accounts/${accountId}/users/${user.username}`" class="btn">
              View
            </router-link>
          </td>
        </tr>
      </tbody>
    </table>
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
nav {
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

nav a {
  color: #3498db;
  text-decoration: none;
}

nav a:hover {
  text-decoration: underline;
}

code {
  background: #f8f8f8;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.8rem;
}

.status-success {
  color: #28a745;
  font-weight: 500;
}

.status-warning {
  color: #ffc107;
  font-weight: 500;
}

@media (max-width: 768px) {
  .table th:nth-child(2),
  .table td:nth-child(2) {
    display: none;
  }
}
</style>