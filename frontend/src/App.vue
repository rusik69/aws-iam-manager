<template>
  <div id="app" :class="{ dark: isDarkTheme }">
    <header>
      <div class="container">
        <div class="logo">
          <router-link to="/" class="logo-link">
            <h1>AWS IAM Manager</h1>
          </router-link>
        </div>
        <nav>
          <router-link to="/" class="nav-link">Users</router-link>
          <router-link to="/public-ips" class="nav-link">Public IPs</router-link>
          <button @click="toggleTheme" class="theme-btn">
            {{ isDarkTheme ? '☀️' : '🌙' }}
          </button>
        </nav>
      </div>
    </header>
    <main>
      <router-view />
    </main>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      isDarkTheme: false
    }
  },
  methods: {
    toggleTheme() {
      this.isDarkTheme = !this.isDarkTheme
      localStorage.setItem('theme', this.isDarkTheme ? 'dark' : 'light')
      this.updateTheme()
    },
    updateTheme() {
      if (this.isDarkTheme) {
        document.documentElement.setAttribute('data-theme', 'dark')
      } else {
        document.documentElement.setAttribute('data-theme', 'light')
      }
    },
    initTheme() {
      const savedTheme = localStorage.getItem('theme')
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      
      this.isDarkTheme = savedTheme === 'dark' || (!savedTheme && prefersDark)
      this.updateTheme()
    }
  },
  mounted() {
    this.initTheme()
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  /* Colors - Light Theme */
  --color-bg-primary: #ffffff;
  --color-bg-secondary: #f8f9fa;
  --color-bg-tertiary: #e9ecef;
  --color-text-primary: #1f2937;
  --color-text-secondary: #6b7280;
  --color-text-tertiary: #9ca3af;
  --color-border: #e5e7eb;
  --color-border-light: #f3f4f6;
  
  /* Button Colors */
  --color-btn-primary: #3b82f6;
  --color-btn-primary-hover: #2563eb;
  --color-btn-secondary: #6b7280;
  --color-btn-secondary-hover: #4b5563;
  --color-btn-success: #10b981;
  --color-btn-success-hover: #059669;
  --color-btn-warning: #f59e0b;
  --color-btn-warning-hover: #d97706;
  --color-btn-danger: #ef4444;
  --color-btn-danger-hover: #dc2626;
  
  /* Status Colors */
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-danger: #ef4444;
  
  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  
  /* Border Radius */
  --radius-sm: 0.375rem;
  --radius: 0.5rem;
  --radius-md: 0.75rem;
  --radius-lg: 1rem;
  --radius-xl: 1.5rem;
  
  /* Spacing */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-2xl: 3rem;
  
  /* Transitions */
  --transition-fast: 150ms ease-in-out;
  --transition: 200ms ease-in-out;
  --transition-slow: 300ms ease-in-out;
}

[data-theme="dark"] {
  /* Colors - Dark Theme */
  --color-bg-primary: #0d1117;
  --color-bg-secondary: #161b22;
  --color-bg-tertiary: #21262d;
  --color-text-primary: #f0f6fc;
  --color-text-secondary: #8b949e;
  --color-text-tertiary: #6e7681;
  --color-border: #30363d;
  --color-border-light: #21262d;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  line-height: 1.6;
  transition: background-color var(--transition), color var(--transition);
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

header {
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  box-shadow: var(--shadow);
  border-bottom: 1px solid var(--color-border);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo-link {
  color: var(--color-text-primary);
  text-decoration: none;
  transition: opacity var(--transition-fast);
}

.logo-link:hover {
  opacity: 0.8;
}

.logo h1 {
  font-size: 1.5rem;
  font-weight: 600;
}

nav {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-link {
  color: var(--color-text-primary);
  text-decoration: none;
  padding: var(--spacing-sm) var(--spacing);
  border-radius: var(--radius);
  transition: all var(--transition-fast);
  font-weight: 500;
  position: relative;
}

.nav-link:hover {
  background: var(--color-bg-secondary);
  color: var(--color-btn-primary);
}

.nav-link.router-link-active {
  background: var(--color-btn-primary);
  color: white;
}

.theme-btn {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  padding: var(--spacing-sm) var(--spacing);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 1.2rem;
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.5rem;
  height: 2.5rem;
}

.theme-btn:hover {
  background: var(--color-bg-tertiary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

main {
  flex: 1;
  padding: var(--spacing-xl);
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  background: var(--color-bg-secondary);
  min-height: calc(100vh - 4rem);
}

.card {
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xl);
  margin-bottom: var(--spacing);
  box-shadow: var(--shadow);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-fast);
}

.card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

/* Modern Button System */
.btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing);
  border-radius: var(--radius);
  border: 1px solid transparent;
  font-size: 0.875rem;
  font-weight: 500;
  line-height: 1.5;
  text-decoration: none;
  cursor: pointer;
  transition: all var(--transition-fast);
  position: relative;
  overflow: hidden;
  
  /* Primary Button */
  background: var(--color-btn-primary);
  color: white;
  box-shadow: var(--shadow-sm);
}

.btn:hover {
  background: var(--color-btn-primary-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow);
}

.btn:active {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* Button Variants */
.btn-secondary {
  background: var(--color-btn-secondary);
  color: white;
}

.btn-secondary:hover {
  background: var(--color-btn-secondary-hover);
}

.btn-success {
  background: var(--color-btn-success);
  color: white;
}

.btn-success:hover {
  background: var(--color-btn-success-hover);
}

.btn-warning {
  background: var(--color-btn-warning);
  color: white;
}

.btn-warning:hover {
  background: var(--color-btn-warning-hover);
}

.btn-danger {
  background: var(--color-btn-danger);
  color: white;
}

.btn-danger:hover {
  background: var(--color-btn-danger-hover);
}

/* Button Sizes */
.btn-sm {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: 0.75rem;
}

.btn-lg {
  padding: var(--spacing) var(--spacing-lg);
  font-size: 1rem;
}

/* Button with Icons */
.btn-icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
}

.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow);
  border: 1px solid var(--color-border-light);
}

.table th, .table td {
  padding: var(--spacing);
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.table th {
  background: var(--color-bg-secondary);
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.table tbody tr {
  transition: background-color var(--transition-fast);
}

.table tbody tr:hover {
  background: var(--color-bg-secondary);
}

.loading {
  text-align: center;
  padding: var(--spacing-2xl);
  color: var(--color-text-secondary);
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing);
}

.loading::before {
  content: '';
  display: block;
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid var(--color-border);
  border-top: 3px solid var(--color-btn-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error {
  background: rgba(239, 68, 68, 0.1);
  color: var(--color-danger);
  padding: var(--spacing);
  border-radius: var(--radius);
  margin: var(--spacing) 0;
  border: 1px solid rgba(239, 68, 68, 0.2);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.error::before {
  content: '⚠️';
  font-size: 1.2rem;
}

@media (max-width: 768px) {
  .container {
    flex-direction: column;
    gap: 1rem;
  }
  
  main {
    padding: 1rem;
  }
  
  .table {
    font-size: 0.875rem;
  }
  
  .table th, .table td {
    padding: 0.5rem;
  }
}
</style>