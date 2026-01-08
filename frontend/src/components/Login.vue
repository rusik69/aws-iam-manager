<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>Cloud IAM Manager</h1>
        <p>Sign in to continue</p>
      </div>
      <form @submit.prevent="handleLogin" class="login-form">
        <div v-if="error" class="error-message" role="alert">
          <strong>Error:</strong> {{ error }}
        </div>
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            required
            autocomplete="username"
            placeholder="Enter your username"
          />
        </div>
        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            autocomplete="current-password"
            placeholder="Enter your password"
          />
        </div>
        <button type="submit" class="login-button" :disabled="loading">
          <span v-if="loading">Signing in...</span>
          <span v-else>Sign In</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Login',
  data() {
    return {
      username: '',
      password: '',
      error: '',
      loading: false
    }
  },
  methods: {
    async handleLogin() {
      this.error = ''
      this.loading = true

      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include',
          body: JSON.stringify({
            username: this.username,
            password: this.password
          })
        })

        let data
        try {
          data = await response.json()
        } catch (jsonErr) {
          // If JSON parsing fails, it's likely a server error
          throw new Error(`Server error: ${response.status} ${response.statusText}`)
        }

        if (!response.ok) {
          const errorMsg = data.message || data.error || `Login failed (${response.status})`
          this.error = errorMsg
          this.loading = false
          return
        }

        // Login successful, wait a moment for cookie to be set, then redirect
        setTimeout(() => {
          this.$router.push('/')
        }, 100)
      } catch (err) {
        // Handle network errors, CORS errors, etc.
        const errorMsg = err.message || 'An error occurred during login. Please check your connection and try again.'
        this.error = errorMsg
        console.error('Login error:', err)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-secondary);
  padding: 2rem;
}

.login-card {
  background: var(--color-bg-primary);
  border-radius: 12px;
  padding: 3rem;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.login-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.login-header p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.95rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 500;
  color: var(--color-text-primary);
  font-size: 0.9rem;
}

.form-group input {
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.error-message {
  padding: 0.75rem 1rem;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-radius: 8px;
  border: 1px solid rgba(239, 68, 68, 0.2);
  font-size: 0.9rem;
  display: block;
  word-wrap: break-word;
}

.error-message strong {
  font-weight: 600;
}

.login-button {
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, var(--color-primary) 0%, #2563eb 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.login-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.login-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}
</style>
