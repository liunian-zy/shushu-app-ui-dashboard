import React from "react";
import ReactDOM from "react-dom/client";
import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import App from "./App";
import { AuthProvider } from "./contexts/AuthContext";
import "./styles/global.css";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Root element not found");
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          fontFamily: "Space Grotesk, PingFang SC, Microsoft Yahei, sans-serif",
          colorPrimary: "#d47f45",
          colorText: "#1d1a17",
          colorBgContainer: "rgba(255,255,255,0.86)",
          borderRadius: 12
        }
      }}
    >
      <AuthProvider>
        <App />
      </AuthProvider>
    </ConfigProvider>
  </React.StrictMode>
);
