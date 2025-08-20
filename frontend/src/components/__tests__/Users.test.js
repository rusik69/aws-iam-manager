import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Users from '../Users.vue'
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

describe('Users.vue', () => {
  let wrapper

  beforeEach(async () => {
    // Mock axios get method
    mockedAxios.get.mockResolvedValue({
      data: [
        {
          username: 'testuser1',
          user_id: 'AIDACKCEVSQ6C2EXAMPLE',
          arn: 'arn:aws:iam::123456789012:user/testuser1',
          create_date: '2023-01-01T00:00:00Z',
          password_set: true,
          access_keys: [
            {
              access_key_id: 'AKIAIOSFODNN7EXAMPLE',
              status: 'Active',
              create_date: '2023-01-01T00:00:00Z'
            }
          ]
        },
        {
          username: 'testuser2',
          user_id: 'AIDACKCEVSQ6C2EXAMPLE2',
          arn: 'arn:aws:iam::123456789012:user/testuser2',
          create_date: '2023-01-01T00:00:00Z',
          password_set: false,
          access_keys: []
        }
      ]
    })

    wrapper = mount(Users, {
      props: {
        accountId: '123456789012'
      },
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

  it('renders users list when data is loaded', async () => {
    expect(wrapper.find('h2').text()).toBe('IAM Users')
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)
  })

  it('displays user information correctly', async () => {
    const firstUserRow = wrapper.findAll('tbody tr')[0]
    const cells = firstUserRow.findAll('td')
    expect(cells[0].text()).toBe('testuser1')
    expect(cells[1].text()).toBe('AIDACKCEVSQ6C2EXAMPLE')
    expect(cells[2].text()).toBe('Yes')
    expect(cells[3].text()).toBe('1')
  })

  it('shows password status correctly', async () => {
    const userRows = wrapper.findAll('tbody tr')
    const firstUserCells = userRows[0].findAll('td')
    const secondUserCells = userRows[1].findAll('td')
    
    expect(firstUserCells[2].text()).toBe('Yes')
    expect(secondUserCells[2].text()).toBe('No')
  })

  it('displays access key count correctly', async () => {
    const userRows = wrapper.findAll('tbody tr')
    const firstUserCells = userRows[0].findAll('td')
    const secondUserCells = userRows[1].findAll('td')
    
    expect(firstUserCells[3].text()).toBe('1')
    expect(secondUserCells[3].text()).toBe('0')
  })

  it('calls API with correct account ID', () => {
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/accounts/123456789012/users')
  })

  it('handles loading state', () => {
    const wrapper = mount(Users, {
      props: {
        accountId: '123456789012'
      },
      global: {
        stubs: {
          'router-link': RouterLinkStub
        }
      }
    })
    
    expect(wrapper.find('.loading').text()).toBe('Loading users...')
  })

  it('handles error state', async () => {
    mockedAxios.get.mockRejectedValueOnce(new Error('API Error'))
    
    const wrapper = mount(Users, {
      props: {
        accountId: '123456789012'
      },
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
    expect(routerLinks.length).toBeGreaterThan(0)
    // First link should be breadcrumb, then user detail links
    expect(routerLinks[1].props('to')).toBe('/accounts/123456789012/users/testuser1')
  })
})