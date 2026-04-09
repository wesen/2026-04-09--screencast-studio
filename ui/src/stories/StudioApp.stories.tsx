import type { Meta, StoryObj } from '@storybook/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { baseApi } from '../api/baseApi';
import { studioDraftReducer } from '../features/studio-draft/studioDraftSlice';
import { sessionReducer } from '../features/session/sessionSlice';
import { StudioApp } from '../components/studio/StudioApp';

const meta = {
  title: 'Studio/StudioApp',
  component: StudioApp,
  tags: ['autodocs'],
  decorators: [
    (Story) => {
      const store = configureStore({
        reducer: {
          [baseApi.reducerPath]: baseApi.reducer,
          studioDraft: studioDraftReducer,
          session: sessionReducer,
        },
        middleware: (getDefaultMiddleware) =>
          getDefaultMiddleware().concat(baseApi.middleware),
      });
      return (
        <Provider store={store}>
          <Story />
        </Provider>
      );
    },
  ],
} satisfies Meta<typeof StudioApp>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
