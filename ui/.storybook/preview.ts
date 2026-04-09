import type { Preview } from '@storybook/react';
import '../src/styles/tokens.css';
import '../src/styles/studio.css';

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    backgrounds: {
      default: 'studio',
      values: [
        {
          name: 'studio',
          value: '#e8e4dc',
        },
        {
          name: 'dark',
          value: '#2c2c2c',
        },
      ],
    },
  },
};

export default preview;
