import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Accounts from '../Accounts.vue'
import axios from 'axios'

// Mock axios
vi.mock('axios')
const mockedAxios = vi.mocked(axios, true)

// Mock router-link
const RouterLinkStub = {
  name: 'RouterLink',
  props: ['to'],
  template: '<a><slot /></a>'
}

describe('Accounts.vue', () => {
  let wrapper

  beforeEach(async () => {
    // Mock axios get method
    mockedAxios.get.mockResolvedValue({
      data: [
        {
          id: '123456789012',
          name: 'Test Account 1'
        },
        {
          id: '123456789013',
          name: 'Test Account 2'
        }
      ]
    })

    wrapper = mount(Accounts, {
      global: {
        stubs: {
          'router-link': RouterLinkStub
        }
      }
    })
    
    // Wait for mounted hook to complete
    await new Promise(resolve => setTimeout(resolve, 0))
    await wrapper.vm.$nextTick()
  })

  it('renders accounts list when data is loaded', async () => {
    expect(wrapper.find('h2').text()).toBe('AWS Accounts')
    expect(wrapper.findAll('.card')).toHaveLength(3) // 1 header card + 2 account cards
  })

  it('displays account information correctly', async () => {
    const firstAccountCard = wrapper.findAll('.card')[1] // Skip header card
    expect(firstAccountCard.find('h3').text()).toBe('Test Account 1')
    expect(firstAccountCard.text()).toContain('123456789012')
  })

  it('calls API to fetch accounts on mount', () => {
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/accounts')
  })

  it('handles loading state initially', () => {
    const wrapper = mount(Accounts, {
      global: {
        stubs: {
          'router-link': RouterLinkStub
        }
      }
    })
    expect(wrapper.find('.loading').text()).toBe('Loading accounts...')
  })

  it('handles error state', async () => {
    mockedAxios.get.mockRejectedValueOnce(new Error('API Error'))
    
    const wrapper = mount(Accounts, {
      global: {
        stubs: {
          'router-link': RouterLinkStub
        }
      }
    })
    
    await new Promise(resolve => setTimeout(resolve, 0))
    await wrapper.vm.$nextTick()
    
    expect(wrapper.find('.error').exists()).toBe(true)
  })

  it('renders router links correctly', async () => {
    const routerLinks = wrapper.findAllComponents(RouterLinkStub)
    expect(routerLinks).toHaveLength(2)
    expect(routerLinks[0].props('to')).toBe('/accounts/123456789012/users')
  })

  it('applies correct CSS classes', async () => {
    const accountCards = wrapper.findAll('.card')
    expect(accountCards[0].classes()).toContain('card')
  })

  it('renders all account data', async () => {
    const accountCards = wrapper.findAll('.card')
    
    // Check first account (index 1, skip header card)
    expect(accountCards[1].text()).toContain('Test Account 1')
    expect(accountCards[1].text()).toContain('123456789012')
    
    // Check second account (index 2)
    expect(accountCards[2].text()).toContain('Test Account 2')
    expect(accountCards[2].text()).toContain('123456789013')
  })
})