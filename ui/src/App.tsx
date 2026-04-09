import React from 'react';
import { Provider } from 'react-redux';
import { store } from '@/app/store';
import { StudioApp } from '@/components/studio/StudioApp';

import '@/styles/tokens.css';
import '@/styles/studio.css';

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <StudioApp />
    </Provider>
  );
};

export default App;
