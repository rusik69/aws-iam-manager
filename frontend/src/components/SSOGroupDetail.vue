<template>
  <div class="sso-group-detail-container">
    <div class="page-header">
      <div class="header-content">
        <button @click="$router.back()" class="btn btn-secondary">
          ← Back
        </button>
        <div class="header-title">
          <h1>{{ group?.display_name || 'SSO Group' }}</h1>
        </div>
        <div class="header-actions">
          <button @click="downloadJSON" class="btn btn-secondary" :disabled="!group">
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
      <p>Loading group details...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <h3>Failed to Load Group</h3>
      <p>{{ error }}</p>
      <button @click="loadData" class="btn btn-primary">Try Again</button>
    </div>

    <div v-else-if="group" class="main-content">
      <div class="group-info-card">
        <h2>Group Information</h2>
        <div class="info-grid">
          <div class="info-item">
            <label>Display Name:</label>
            <span>{{ group.display_name }}</span>
          </div>
          <div class="info-item">
            <label>Group ID:</label>
            <span>{{ group.group_id }}</span>
          </div>
          <div class="info-item">
            <label>Description:</label>
            <span>{{ group.description || '-' }}</span>
          </div>
          <div class="info-item">
            <label>Members:</label>
            <span>{{ members.length }}</span>
          </div>
        </div>
      </div>

      <div class="members-card">
        <h2>Group Members ({{ members.length }})</h2>
        <div v-if="members.length === 0" class="empty-state">
          <p>No members found</p>
        </div>
        <table v-else class="members-table">
          <thead>
            <tr>
              <th>Type</th>
              <th>Name</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="member in members" :key="member.member_id">
              <td>{{ member.member_type }}</td>
              <td>{{ member.display_name || member.user_name || member.member_id }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="assignments-card">
        <h2>Account Assignments ({{ assignments.length }})</h2>
        <div v-if="assignments.length === 0" class="empty-state">
          <p>No account assignments found</p>
        </div>
        <table v-else class="assignments-table">
          <thead>
            <tr>
              <th>Account ID</th>
              <th>Account Name</th>
              <th>Permission Set</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="assignment in assignments" :key="`${assignment.account_id}-${assignment.permission_set_arn}`">
              <td>{{ assignment.account_id }}</td>
              <td>{{ assignment.account_name || '-' }}</td>
              <td>{{ assignment.permission_set_name || assignment.permission_set_arn }}</td>
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
  name: 'SSOGroupDetail',
  props: {
    groupId: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      group: null,
      members: [],
      assignments: [],
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
        const [groupResponse, membersResponse, assignmentsResponse] = await Promise.all([
          axios.get(`/api/sso/groups/${this.groupId}`),
          axios.get(`/api/sso/groups/${this.groupId}/members`),
          axios.get(`/api/sso/groups/${this.groupId}/assignments`)
        ])
        this.group = groupResponse.data
        this.members = membersResponse.data
        this.assignments = assignmentsResponse.data
      } catch (err) {
        this.error = err.response?.data?.error || err.message || 'Failed to load group details'
      } finally {
        this.loading = false
      }
    },
    downloadJSON() {
      const data = {
        group: this.group,
        members: this.members,
        assignments: this.assignments
      }
      const dataStr = JSON.stringify(data, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `sso-group-${this.group?.group_id || 'details'}-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    }
  }
}
</script>

<style scoped>
.sso-group-detail-container {
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

.group-info-card, .members-card, .assignments-card {
  background: var(--color-bg-primary);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid var(--color-border);
}

.group-info-card h2, .members-card h2, .assignments-card h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
  color: var(--color-text-primary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
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

.members-table, .assignments-table {
  width: 100%;
  border-collapse: collapse;
}

.members-table thead, .assignments-table thead {
  background: var(--color-bg-secondary);
}

.members-table th, .assignments-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-primary);
  border-bottom: 2px solid var(--color-border);
}

.members-table tbody tr, .assignments-table tbody tr {
  background: var(--color-bg-primary);
}

.members-table td, .assignments-table td {
  padding: 12px;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-primary);
  background: var(--color-bg-primary);
}

.members-table tbody tr:hover, .assignments-table tbody tr:hover {
  background: var(--color-bg-secondary);
}

.members-table tbody tr:hover td, .assignments-table tbody tr:hover td {
  background: var(--color-bg-secondary);
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
</style>
