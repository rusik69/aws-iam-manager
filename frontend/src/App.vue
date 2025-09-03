<template>
  <div id="app" :class="{ dark: isDarkTheme }">
    <header>
      <div class="container">
        <div class="logo">
          <h1>AWS IAM Manager</h1>
        </div>
        <nav>
          <router-link to="/">Accounts</router-link>
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

body {
  font-family: -apple-system, BlinkMacSystemFont, sans-serif;
  background: #f5f5f5;
  color: #333;
  line-height: 1.6;
}

.dark body {
  background: #1a1a1a;
  color: #e0e0e0;
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

header {
  background: #2c3e50;
  color: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.dark header {
  background: #111;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
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

nav a {
  color: white;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  transition: background 0.2s;
}

nav a:hover {
  background: rgba(255,255,255,0.1);
}

.theme-btn {
  background: none;
  border: none;
  color: white;
  padding: 0.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 1.2rem;
  transition: background 0.2s;
}

.theme-btn:hover {
  background: rgba(255,255,255,0.1);
}

main {
  flex: 1;
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.card {
  background: white;
  border-radius: 0.5rem;
  padding: 1.5rem;
  margin-bottom: 1rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.dark .card {
  background: #2a2a2a;
}

.btn {
  background: #3498db;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
  font-size: 0.875rem;
  transition: background 0.2s;
  margin: 0.25rem;
}

.btn:hover {
  background: #2980b9;
}

.btn-danger {
  background: #e74c3c;
}

.btn-danger:hover {
  background: #c0392b;
}

.btn-secondary {
  background: #95a5a6;
}

.btn-secondary:hover {
  background: #7f8c8d;
}

.table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 0.5rem;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.dark .table {
  background: #2a2a2a;
}

.table th, .table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.dark .table th, .dark .table td {
  border-bottom: 1px solid #444;
}

.table th {
  background: #f8f9fa;
  font-weight: 600;
}

.dark .table th {
  background: #333;
}

.table tbody tr:hover {
  background: #f8f9fa;
}

.dark .table tbody tr:hover {
  background: #333;
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.loading::before {
  content: '';
  display: block;
  width: 2rem;
  height: 2rem;
  border: 2px solid #ddd;
  border-top: 2px solid #3498db;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error {
  background: #fee;
  color: #c00;
  padding: 1rem;
  border-radius: 0.25rem;
  margin: 1rem 0;
  border: 1px solid #fcc;
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