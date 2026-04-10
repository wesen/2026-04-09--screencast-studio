import { baseApi } from './baseApi';
import { decodeProto } from './proto';
import type { DiscoveryResponse, HealthResponse } from './types';
import {
  DiscoveryResponseSchema,
  HealthResponseSchema,
} from '@/gen/proto/screencast/studio/v1/web_pb';

export const discoveryApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getHealth: builder.query<HealthResponse, void>({
      query: () => '/healthz',
      transformResponse: (response) => decodeProto(HealthResponseSchema, response),
    }),
    getDiscovery: builder.query<DiscoveryResponse, void>({
      query: () => '/discovery',
      transformResponse: (response) => decodeProto(DiscoveryResponseSchema, response),
      providesTags: ['Discovery'],
    }),
  }),
});

export const { useGetDiscoveryQuery, useGetHealthQuery } = discoveryApi;
