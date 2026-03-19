import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <SettingsProvider>
            <App/>
        </SettingsProvider>
    </React.StrictMode>
)
