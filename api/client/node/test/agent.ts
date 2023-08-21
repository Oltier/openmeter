import { MockAgent } from 'undici'
import { mockMeter, mockMeterValue } from './mocks.js'

export const mockAgent = new MockAgent()
mockAgent.disableNetConnect()

const client = mockAgent.get('http://127.0.0.1:8888')
client
  .intercept({
    path: '/api/v1/events',
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/cloudevents+json',
    },
    body: JSON.stringify({
      specversion: '1.0',
      id: 'id-1',
      source: 'my-app',
      type: 'my-type',
      subject: 'my-awesome-user-id',
      time: new Date('2023-01-01'),
      data: {
        api_calls: 1,
      },
    }),
  })
  .reply(204)

client
  .intercept({
    path: '/api/v1/events',
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/cloudevents+json',
    },
    body: JSON.stringify({
      specversion: '1.0',
      id: 'aaf17be7-860c-4519-91d3-00d97da3cc65',
      source: '@openmeter/sdk',
      type: 'my-type',
      subject: 'my-awesome-user-id',
      data: {
        api_calls: 1,
      },
    }),
  })
  .reply(204)

client
  .intercept({
    path: '/api/v1/meters',
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  })
  .reply(200, [mockMeter], {
    headers: {
      'Content-Type': 'application/json',
    },
  })

client
  .intercept({
    path: `/api/v1/meters/${mockMeter.slug}`,
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  })
  .reply(200, mockMeter, {
    headers: {
      'Content-Type': 'application/json',
    },
  })

client
  .intercept({
    path: `/api/v1/meters/${mockMeter.slug}/values`,
    query: {},
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  })
  .reply(
    200,
    {
      windowSize: 'HOUR',
      data: [mockMeterValue],
    },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  )

client
  .intercept({
    path: `/api/v1/meters/${mockMeter.slug}/values`,
    query: {
      subject: 'user-1',
      from: new Date('2021-01-01').toISOString(),
      to: new Date('2021-01-02').toISOString(),
      windowSize: 'HOUR',
    },
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  })
  .reply(
    200,
    {
      windowSize: 'HOUR',
      data: [mockMeterValue],
    },
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  )