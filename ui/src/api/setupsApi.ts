import { baseApi } from './baseApi';
import type { CompileResponse, NormalizeResponse } from './types';

export const setupsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getExample: builder.query<{ dsl: string }, void>({
      query: () => '/setups/example',
      providesTags: ['Setups'],
    }),
    normalizeSetup: builder.mutation<NormalizeResponse, { dsl: string }>({
      query: (body) => ({
        url: '/setups/normalize',
        method: 'POST',
        body: { dsl_format: 'yaml', ...body },
      }),
    }),
    compileSetup: builder.mutation<CompileResponse, { dsl: string }>({
      query: (body) => ({
        url: '/setups/compile',
        method: 'POST',
        body: { dsl_format: 'yaml', ...body },
      }),
      invalidatesTags: ['Setups'],
    }),
    validateSetup: builder.mutation<{ valid: boolean; errors: string[] }, { dsl: string }>({
      query: (body) => ({
        url: '/setups/validate',
        method: 'POST',
        body: { dsl_format: 'yaml', ...body },
      }),
    }),
  }),
});

export const {
  useGetExampleQuery,
  useNormalizeSetupMutation,
  useCompileSetupMutation,
  useValidateSetupMutation,
} = setupsApi;
