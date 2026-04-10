import type { Meta, StoryObj } from '@storybook/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { baseApi } from '../api/baseApi';
import { editorReducer } from '../features/editor/editorSlice';
import { previewReducer } from '../features/previews/previewSlice';
import { setupReducer } from '../features/setup/setupSlice';
import { setupDraftReducer } from '../features/setup-draft/setupDraftSlice';
import { studioUiReducer } from '../features/studio-ui/studioUiSlice';
import { sessionReducer } from '../features/session/sessionSlice';
import { StudioPage } from '../pages/StudioPage';

const meta = {
  title: 'Pages/StudioPage',
  component: StudioPage,
  tags: ['autodocs'],
  decorators: [
    (Story) => {
      const store = configureStore({
        reducer: {
          [baseApi.reducerPath]: baseApi.reducer,
          editor: editorReducer,
          previews: previewReducer,
          setup: setupReducer,
          setupDraft: setupDraftReducer,
          studioUi: studioUiReducer,
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
} satisfies Meta<typeof StudioPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
