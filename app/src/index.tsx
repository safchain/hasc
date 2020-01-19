import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';
import { SnackbarProvider } from 'notistack';

ReactDOM.render(
    <SnackbarProvider maxSnack={1}>
        <App />
    </SnackbarProvider>
    , document.getElementById('root'));
serviceWorker.unregister();
