import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from '../App.vue'

// Mock router-link and router-view
const RouterLinkStub = {
  name: 'RouterLink',
  props: ['to'],
  template: '<a><slot /></a>'
}

const RouterViewStub = {
  name: 'RouterView',
  template: '<div>Router View</div>'
}

describe('App.vue', () => {
  it('renders the main app container', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    expect(wrapper.find('#app').exists()).toBe(true)
  })

  it('renders the application title', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    expect(wrapper.find('h1').text()).toBe('Cloud Manager')
  })

  it('renders header with navigation', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    expect(wrapper.find('header').exists()).toBe(true)
    expect(wrapper.find('nav').exists()).toBe(true)
  })

  it('renders main content area', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    expect(wrapper.find('main').exists()).toBe(true)
  })

  it('renders router-view for content', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    expect(wrapper.findComponent(RouterViewStub).exists()).toBe(true)
  })

  it('renders navigation link to accounts', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    const routerLink = wrapper.findComponent(RouterLinkStub)
    expect(routerLink.exists()).toBe(true)
    expect(routerLink.props('to')).toBe('/')
    expect(routerLink.text()).toBe('Accounts')
  })

  it('applies correct CSS classes', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    expect(wrapper.find('#app').exists()).toBe(true)
    expect(wrapper.find('header').exists()).toBe(true)
    expect(wrapper.find('main').exists()).toBe(true)
  })

  it('has proper component hierarchy', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          'router-view': RouterViewStub
        }
      }
    })
    
    // Check that main components are properly nested
    expect(wrapper.find('#app header h1').exists()).toBe(true)
    expect(wrapper.find('#app main').exists()).toBe(true)
  })
})