import React from 'react';
import { Provider } from 'react-redux';
import { store } from '@/app/store';
import { StudioPage } from '@/pages/StudioPage';

import '@/styles/tokens.css';
import '@/styles/studio.css';

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <StudioPage />
    </Provider>
  );
};

export default App;
