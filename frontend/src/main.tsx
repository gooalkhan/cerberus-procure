import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

const render = () => {
  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  )
}

// 로컬 개발 모드(npm run dev)이거나 GitHub Pages 환경일 때 WASM 로드
const shouldLoadWasm = import.meta.env.DEV || 
                       window.location.hostname.includes('github.io') || 
                       window.location.search.includes('wasm');

if (shouldLoadWasm) {
  const go = new (window as any).Go();
  WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    render();
  });
} else {
  render();
}
