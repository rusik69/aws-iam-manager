import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDetail from '../UserDetail.vue'
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

// Mock window.confirm and window.alert
global.confirm = vi.fn(() => true)
global.alert = vi.fn()

describe('UserDetail.vue', () => {
  let wrapper

  const mockUser = {
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
  }

  beforeEach(async () => {
    // Mock axios methods
    mockedAxios.get.mockResolvedValue({ data: mockUser })
    mockedAxios.post.mockResolvedValue({
      data: {
        access_key_id: 'AKIAIOSFODNN7EXAMPLE2',
        secret_access_key: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
        status: 'Active',
        create_date: '2023-01-01T00:00:00Z'
      }
    })
    mockedAxios.delete.mockResolvedValue({ data: { message: 'Access key deleted successfully' } })
    mockedAxios.put.mockResolvedValue({
      data: {
        access_key_id: 'AKIAIOSFODNN7EXAMPLE3',
        secret_access_key: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY2',
        status: 'Active',
        create_date: '2023-01-01T00:00:00Z',
        message: 'Access key rotated successfully'
      }
    })

    wrapper = mount(UserDetail, {
      props: {
        accountId: '123456789012',
        username: 'testuser1'
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

  it('renders user details when data is loaded', async () => {
    expect(wrapper.find('h2').text()).toBe('User Details: testuser1')
    expect(wrapper.text()).toContain('AIDACKCEVSQ6C2EXAMPLE')
    expect(wrapper.text()).toContain('arn:aws:iam::123456789012:user/testuser1')
  })

  it('displays access keys section', async () => {
    const h3Elements = wrapper.findAll('h3')
    expect(h3Elements[1].text()).toBe('Access Keys')
    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
  })

  it('shows access key information', async () => {
    const accessKeyRow = wrapper.find('tbody tr')
    expect(accessKeyRow.text()).toContain('AKIAIOSFODNN7EXAMPLE')
    expect(accessKeyRow.text()).toContain('Active')
  })

  it('calls API with correct parameters on mount', () => {
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/accounts/123456789012/users/testuser1')
  })

  it('handles loading state', () => {
    const wrapper = mount(UserDetail, {
      props: {
        accountId: '123456789012',
        username: 'testuser1'
      },
      global: {
        stubs: {
          'router-link': RouterLinkStub
        }
      }
    })
    
    expect(wrapper.find('.loading').text()).toBe('Loading user details...')
  })

  it('handles error state', async () => {
    mockedAxios.get.mockRejectedValueOnce(new Error('API Error'))
    
    const wrapper = mount(UserDetail, {
      props: {
        accountId: '123456789012',
        username: 'testuser1'
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

  it('creates new access key when button is clicked', async () => {
    const createButton = wrapper.find('button')
    expect(createButton.text()).toBe('Create New Key')
    
    await createButton.trigger('click')
    
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/accounts/123456789012/users/testuser1/keys')
  })

  it('deletes access key when delete button is clicked', async () => {
    const deleteButton = wrapper.find('.btn-danger')
    await deleteButton.trigger('click')
    
    expect(mockedAxios.delete).toHaveBeenCalledWith('/api/accounts/123456789012/users/testuser1/keys/AKIAIOSFODNN7EXAMPLE')
  })

  it('rotates access key when rotate button is clicked', async () => {
    const rotateButton = wrapper.find('.btn-secondary')
    await rotateButton.trigger('click')
    
    expect(mockedAxios.put).toHaveBeenCalledWith('/api/accounts/123456789012/users/testuser1/keys/AKIAIOSFODNN7EXAMPLE/rotate')
  })

  it('displays password status correctly', async () => {
    expect(wrapper.text()).toContain('Password Set:')
    expect(wrapper.text()).toContain('Yes')
  })

  it('updates user data after creating access key', async () => {
    // Reset the mock call count before the test
    mockedAxios.get.mockClear()
    
    // Trigger initial load by mounting fresh component
    const testWrapper = mount(UserDetail, {
      props: {
        accountId: '123456789012',
        username: 'testuser1'
      },
      global: {
        stubs: {
          'router-link': RouterLinkStub
        }
      }
    })
    
    await new Promise(resolve => setTimeout(resolve, 0))
    await testWrapper.vm.$nextTick()
    
    const createButton = testWrapper.find('button')
    await createButton.trigger('click')
    
    // Wait for the component to update
    await testWrapper.vm.$nextTick()
    
    expect(mockedAxios.get).toHaveBeenCalledTimes(2) // Initial load + refresh after create
  })
})