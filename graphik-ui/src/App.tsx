import React from 'react';
import logo from './logo.svg';
import './App.css';
import { Provider } from 'react-redux'
import { Playground, store } from 'graphql-playground-react'

function App() {
  return (
      <div className="App">
        <Provider store={store}>
          <Playground endpoint='http://localhost:' />
        </Provider>,
      </div>
  );
}

export default App;
