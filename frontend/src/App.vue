<template>
  <div id="app" :class="{ 'dark-theme': isDarkTheme }">
    <header class="header">
      <div class="header-content">
        <div class="header-left">
          <div class="logo">
            <svg class="logo-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2ZM21 9V7L15 1H5C3.89 1 3 1.89 3 3V18A2 2 0 0 0 5 20H11V18H5V3H13V9H21Z"/>
              <path d="M17 13C15.34 13 14 14.34 14 16S15.34 19 17 19 20 17.66 20 16 18.66 13 17 13M17 14.5C18.11 14.5 19 15.39 19 16.5S18.11 18.5 17 18.5 15 17.61 15 16.5 15.89 14.5 17 14.5M24 16L22.5 14.5L17 20L14.5 17.5L13 19L17 23L24 16Z"/>
            </svg>
            <h1>AWS IAM Manager</h1>
          </div>
        </div>
        <nav class="nav">
          <router-link to="/" class="nav-link">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M10 20V14H14V20H19V12H22L12 3L2 12H5V20H10Z"/>
            </svg>
            Accounts
          </router-link>
          <button @click="toggleTheme" class="theme-toggle" :title="isDarkTheme ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
            <svg v-if="isDarkTheme" class="theme-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z"/>
            </svg>
            <svg v-else class="theme-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z"/>
            </svg>
          </button>
        </nav>
      </div>
    </header>
    <main class="main">
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
:root {
  /* Light theme colors */
  --color-bg-primary: #ffffff;
  --color-bg-secondary: #f8f9fa;
  --color-bg-tertiary: #e9ecef;
  --color-bg-accent: #007bff;
  --color-text-primary: #212529;
  --color-text-secondary: #6c757d;
  --color-text-tertiary: #adb5bd;
  --color-text-inverse: #ffffff;
  --color-border: #dee2e6;
  --color-border-light: #e9ecef;
  --color-shadow: rgba(0, 0, 0, 0.1);
  --color-shadow-light: rgba(0, 0, 0, 0.05);
  
  /* Status colors */
  --color-success: #28a745;
  --color-danger: #dc3545;
  --color-warning: #ffc107;
  --color-info: #17a2b8;
  
  /* Button colors */
  --color-btn-primary: #007bff;
  --color-btn-primary-hover: #0056b3;
  --color-btn-secondary: #6c757d;
  --color-btn-secondary-hover: #545b62;
  --color-btn-danger: #dc3545;
  --color-btn-danger-hover: #c82333;
  --color-btn-warning: #ffc107;
  --color-btn-warning-hover: #e0a800;
  
  /* Header colors */
  --color-header-bg: #343a40;
  --color-header-text: #ffffff;
  --color-header-accent: #007bff;
  
  /* Spacing */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-xxl: 3rem;
  
  /* Border radius */
  --radius-sm: 0.25rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
  --radius-xl: 1rem;
  
  /* Animations */
  --transition-fast: 0.15s ease-in-out;
  --transition-normal: 0.2s ease-in-out;
  --transition-slow: 0.3s ease-in-out;
}

[data-theme="dark"] {
  /* Dark theme colors */
  --color-bg-primary: #1a1a1a;
  --color-bg-secondary: #2d3748;
  --color-bg-tertiary: #4a5568;
  --color-bg-accent: #3182ce;
  --color-text-primary: #f7fafc;
  --color-text-secondary: #cbd5e0;
  --color-text-tertiary: #a0aec0;
  --color-text-inverse: #1a202c;
  --color-border: #4a5568;
  --color-border-light: #2d3748;
  --color-shadow: rgba(0, 0, 0, 0.3);
  --color-shadow-light: rgba(0, 0, 0, 0.15);
  
  /* Header colors for dark theme */
  --color-header-bg: #1a202c;
  --color-header-text: #f7fafc;
  --color-header-accent: #3182ce;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
  background-color: var(--color-bg-secondary);
  color: var(--color-text-primary);
  line-height: 1.6;
  transition: background-color var(--transition-normal), color var(--transition-normal);
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Header Styles */
.header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  box-shadow: 0 2px 8px var(--color-shadow);
  position: sticky;
  top: 0;
  z-index: 100;
  transition: background-color var(--transition-normal);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-md) var(--spacing-xl);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.logo-icon {
  width: 2rem;
  height: 2rem;
  color: var(--color-header-accent);
}

.logo h1 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--color-header-text);
}

.nav {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.nav-link {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--color-header-text);
  text-decoration: none;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  transition: all var(--transition-normal);
  font-weight: 500;
}

.nav-link:hover {
  background-color: rgba(255, 255, 255, 0.1);
  transform: translateY(-1px);
}

.nav-icon {
  width: 1.2rem;
  height: 1.2rem;
}

.theme-toggle {
  background: none;
  border: none;
  color: var(--color-header-text);
  padding: var(--spacing-sm);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-toggle:hover {
  background-color: rgba(255, 255, 255, 0.1);
  transform: rotate(15deg);
}

.theme-icon {
  width: 1.5rem;
  height: 1.5rem;
}

/* Main Content */
.main {
  flex: 1;
  padding: var(--spacing-xl);
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

/* Card Styles */
.card {
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
  box-shadow: 0 4px 12px var(--color-shadow-light);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-normal);
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px var(--color-shadow);
}

/* Button Styles */
.btn {
  background: var(--color-btn-primary);
  color: var(--color-text-inverse);
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  margin-right: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
  font-weight: 500;
  font-size: 0.875rem;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.btn:hover {
  background: var(--color-btn-primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px var(--color-shadow);
}

.btn:active {
  transform: translateY(0);
}

.btn-danger {
  background: var(--color-btn-danger);
}

.btn-danger:hover {
  background: var(--color-btn-danger-hover);
}

.btn-secondary {
  background: var(--color-btn-secondary);
}

.btn-secondary:hover {
  background: var(--color-btn-secondary-hover);
}

.btn-warning {
  background: var(--color-btn-warning);
  color: var(--color-text-primary);
}

.btn-warning:hover {
  background: var(--color-btn-warning-hover);
}

/* Table Styles */
.table {
  width: 100%;
  border-collapse: collapse;
  margin-top: var(--spacing-md);
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: 0 2px 8px var(--color-shadow-light);
}

.table th,
.table td {
  padding: var(--spacing-md);
  text-align: left;
  border-bottom: 1px solid var(--color-border-light);
}

.table th {
  background-color: var(--color-bg-secondary);
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.table tbody tr {
  transition: background-color var(--transition-fast);
}

.table tbody tr:hover {
  background-color: var(--color-bg-secondary);
}

.table tbody tr:last-child td {
  border-bottom: none;
}

/* Status Styles */
.status-active {
  color: var(--color-success);
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.status-inactive {
  color: var(--color-danger);
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.status-active::before,
.status-inactive::before {
  content: '';
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: currentColor;
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
  margin-bottom: var(--spacing-md);
  border: 1px solid rgba(220, 53, 69, 0.2);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.error::before {
  content: '⚠';
  font-size: 1.2rem;
}

/* Breadcrumb */
.breadcrumb {
  margin-bottom: var(--spacing-md);
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.breadcrumb a {
  color: var(--color-btn-primary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.breadcrumb a:hover {
  color: var(--color-btn-primary-hover);
  text-decoration: underline;
}

/* Responsive Design */
@media (max-width: 768px) {
  .header-content {
    padding: var(--spacing-md);
    flex-direction: column;
    gap: var(--spacing-md);
  }
  
  .logo h1 {
    font-size: 1.25rem;
  }
  
  .main {
    padding: var(--spacing-md);
  }
  
  .table {
    font-size: 0.875rem;
  }
  
  .table th,
  .table td {
    padding: var(--spacing-sm);
  }
}

@media (max-width: 480px) {
  .nav {
    flex-direction: column;
    width: 100%;
  }
  
  .card {
    padding: var(--spacing-md);
  }
  
  .btn {
    font-size: 0.75rem;
    padding: var(--spacing-xs) var(--spacing-sm);
  }
}

/* Dark theme specific adjustments */
.dark-theme .table tbody tr:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.dark-theme .nav-link:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.dark-theme .theme-toggle:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
</style>