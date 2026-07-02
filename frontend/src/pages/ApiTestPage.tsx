import React, { useState } from 'react';
import axios from 'axios';
import { apiClient } from '@/api';
import { ProductCatalogPage } from '@/pages/ProductCatalogPage';
import { AccessControlPage } from '@/pages/AccessControlPage';
import { ProductRoleConsolePage } from '@/pages/ProductRoleConsolePage';
import { useAuthStore } from '@/store';

export function ApiTestPage() {
  const [view, setView] = useState<'console' | 'catalog' | 'grant' | 'roles' | 'sensitive'>('console');
  

  // Global settings
  const [baseURL, setBaseURL] = useState('http://localhost:2910');
  const { accessToken, clearAuth } = useAuthStore();
  const [manualToken, setManualToken] = useState(accessToken || '');
  

  // Sync manualToken with the store's accessToken when it changes (e.g. after login)
  React.useEffect(() => {
    if (accessToken) {
      setManualToken(accessToken);
    }
  }, [accessToken]);

  React.useEffect(() => {
    const trimmed = baseURL.trim().replace(/\/$/, '');
    const normalizedBaseURL = trimmed.endsWith('/api/v1')
      ? trimmed
      : trimmed.endsWith('/api')
        ? `${trimmed}/v1`
        : `${trimmed}/api/v1`;

    apiClient.defaults.baseURL = normalizedBaseURL;
  }, [baseURL]);

  React.useEffect(() => {
    if (manualToken) {
      apiClient.defaults.headers.common.Authorization = `Bearer ${manualToken}`;
      return;
    }

    delete apiClient.defaults.headers.common.Authorization;
  }, [manualToken]);

  // Log panels for displaying API logs
  const [apiLogs, setApiLogs] = useState<Array<{
    action: string;
    method: string;
    url: string;
    payload?: any;
    status?: number;
    response?: any;
    error?: string;
    timestamp: string;
  }>>([]);

  // Form states
  const [registerForm, setRegisterForm] = useState({
    email: 'testuser_' + Math.floor(Math.random() * 100000) + '@example.com',
    username: 'testuser_' + Math.floor(Math.random() * 100000),
    password: 'Password123!',
    confirm_password: 'Password123!',
    first_name: 'Test',
    last_name: 'User',
  });

  const [loginForm, setLoginForm] = useState({
    identifier: '',
    password: 'Password123!',
  });

  const [giveAdminForm, setGiveAdminForm] = useState({
    id: '',
    role_id: '1', // 1: GOD, 2: Owner, 3: Superadmin, 4: User (จาก bootstrap seed)
  });

  const [productForm, setProductForm] = useState({
    product_name: 'Product_' + Math.floor(Math.random() * 100),
    product_code: 'PC' + Math.floor(Math.random() * 100),
    env_code: 'DEV',
    env_name: 'Development',
  });

   const [ingestForm, setIngestForm] = useState({
    productId: '1',
    environmentId: '1',
    projectId: '',
    categoryId: '',
    logLevel: 'INFO',
    message: 'สวัสดีจากระบบทดสอบ Log Ingest!',
    testMode: 'NORMAL', // เพิ่มค่านี้ (NORMAL, RETRY_THEN_SUCCESS, ALWAYS_RETRY, PERMANENT_FAIL)
  });


  const [loading, setLoading] = useState<Record<string, boolean>>({});

  // Sensitive logs testing form states
  const [sensitiveProductId, setSensitiveProductId] = useState('1');
  const [sensitiveLogId, setSensitiveLogId] = useState('');
  const [sensitiveFieldPath, setSensitiveFieldPath] = useState('');
  const [sensitiveReason, setSensitiveReason] = useState('ต้องการดูข้อมูลเชิงลึกสำหรับใช้ตรวจสอบระบบ');
  const [sensitiveRequestId, setSensitiveRequestId] = useState('');
  const [sensitiveApprovalStatus, setSensitiveApprovalStatus] = useState('REJECTED');
  const [sensitiveRevealPassword, setSensitiveRevealPassword] = useState('Password123!');

  // Helper to add logs to the log display
  const addLog = (action: string, method: string, url: string, payload: any, status?: number, response?: any, error?: string) => {
    setApiLogs((prev) => [
      {
        action,
        method,
        url,
        payload,
        status,
        response,
        error,
        timestamp: new Date().toLocaleTimeString(),
      },
      ...prev,
    ]);
  };

  // Helper axios client config
  const getClient = (useToken = false) => {
    const tokenToUse = manualToken || accessToken;
    return axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
        ...(useToken && tokenToUse ? { 'Authorization': `Bearer ${tokenToUse}` } : {}),
      },
    });
  };

  // 1. Register handler
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    const actionName = 'AUTH.REGISTER.CREATE';
    const path = '/api/v1/auth/register';
    setLoading((prev) => ({ ...prev, register: true }));

    try {
      const res = await getClient(false).post(path, registerForm);
      addLog(actionName, 'POST', path, registerForm, res.status, res.data);
      // Auto fill register email to login form for convenience
      setLoginForm((prev) => ({ ...prev, identifier: registerForm.email }));
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'POST', path, registerForm, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, register: false }));
    }
  };

  // 2. Login handler
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    const actionName = 'AUTH.LOGIN.CREATE';
    const path = '/api/v1/auth/login';
    setLoading((prev) => ({ ...prev, login: true }));

    try {
      const res = await getClient(false).post(path, loginForm);
      const token = res.data?.data?.access_token || res.data?.data?.accessToken || res.data?.accessToken;
      const role = res.data?.data?.role || res.data?.role;
      const userId = res.data?.data?.user_id || res.data?.user_id;

      if (token) {
        setManualToken(token);
        // Save to Zustand store to keep session across the app safely
        useAuthStore.setState({
          accessToken: token,
          isAuthenticated: true,
          roles: role ? [role] : [],
          currentUser: res.data?.data?.user || res.data?.user || { 
            id: String(userId || ''), 
            username: loginForm.identifier.split('@')[0],
            email: loginForm.identifier,
            fullName: 'Platform User',
            status: 'active',
            roles: role ? [{ id: '0', name: role, permissions: [] }] : [],
            permissions: [],
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        });

        // Auto fill user_id to giveAdminForm for convenience
        if (userId) {
          setGiveAdminForm((prev) => ({ ...prev, id: String(userId) }));
        }
      }
      addLog(actionName, 'POST', path, loginForm, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'POST', path, loginForm, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, login: false }));
    }
  };

  // 3. Give Admin Access handler
  const handleGiveAdmin = async (e: React.FormEvent) => {
    e.preventDefault();
    const actionName = 'ADMIN.GIVE_ADMIN.UPDATE';
    const path = '/api/v1/admin/give-admin';
    setLoading((prev) => ({ ...prev, giveAdmin: true }));

    const payload = {
      id: parseInt(giveAdminForm.id, 10),
      role_id: parseInt(giveAdminForm.role_id, 10),
    };

    try {
      const res = await getClient(true).put(path, payload);
      addLog(actionName, 'PUT', path, payload, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'PUT', path, payload, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, giveAdmin: false }));
    }
  };

  // 4. Create Product handler
  const handleCreateProduct = async (e: React.FormEvent) => {
    e.preventDefault();
    const actionName = 'PRODUCTS.CREATE';
    const path = '/api/v1/products';
    setLoading((prev) => ({ ...prev, product: true }));

    const payload = {
      product_name: productForm.product_name,
      product_code: productForm.product_code,
      environments: [
        {
          environment_code: productForm.env_code,
          environment_name: productForm.env_name,
        },
      ],
    };

    try {
      const res = await getClient(true).post(path, payload);
      addLog(actionName, 'POST', path, payload, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'POST', path, payload, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, product: false }));
    }
  };

  // Log Ingestion Handler
  const handleIngestLog = async (e: React.FormEvent) => {
    e.preventDefault();
    const productId = parseInt(ingestForm.productId, 10);
    const envId = parseInt(ingestForm.environmentId, 10);
    const projId = ingestForm.projectId ? parseInt(ingestForm.projectId, 10) : null;
    const catId = ingestForm.categoryId ? parseInt(ingestForm.categoryId, 10) : null;

    // กำหนด message ตาม Test Mode ที่เลือก
    let testMessage = ingestForm.message;
    if (ingestForm.testMode === 'RETRY_THEN_SUCCESS') {
      testMessage = 'FORCE_RETRY_THEN_SUCCESS';
    } else if (ingestForm.testMode === 'ALWAYS_RETRY') {
      testMessage = 'FORCE_RETRY';
    } else if (ingestForm.testMode === 'PERMANENT_FAIL') {
      testMessage = 'FORCE_FAIL';
    }

    const actionName1 = 'QUEUE.INGEST.CREATE';
    const path1 = '/api/v1/queues';
    setLoading((prev) => ({ ...prev, ingest: true }));

    const payload = {
      product_id: productId,
      environment_id: envId,
      queue_key: 'default',
      source_type: 'application',
      source_platform: 'backend',
      priority: 1,
      logs: [
        {
          sequence_no: 1,
          source_type: 'application',
          source_platform: 'backend',
          input_payload: {
            log_level: ingestForm.logLevel,
            event_type: 'MANUAL_TEST',
            message: testMessage, // ใช้ testMessage แทน
            timestamp: new Date().toISOString(),
            ...(projId ? { project_id: projId } : {}),
            ...(catId ? { category_id: catId } : {}),
          },
        },
      ],
    };
    try {
      const res = await getClient(true).post(path1, payload);
      addLog(actionName1, 'POST', path1, payload, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName1, 'POST', path1, payload, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, ingest: false }));
    }
  };
  // 5. Create Sensitive Access Request Handler
  const handleCreateSensitiveRequest = async (e: React.FormEvent) => {
    e.preventDefault();
    const productId = parseInt(sensitiveProductId, 10);
    const actionName = 'SENSITIVE_LOGS.REQUEST.CREATE';
    const path = `/api/v1/products/${productId}/sensitive-logs/requests`;
    setLoading((prev) => ({ ...prev, sensitiveRequest: true }));

    const payload: any = {
      product_id: productId,
      reason: sensitiveReason,
    };
    if (sensitiveLogId) payload.log_id = sensitiveLogId;
    if (sensitiveFieldPath) payload.field_path = sensitiveFieldPath;

    try {
      const res = await getClient(true).post(path, payload);
      if (res.data?.data?.request_id) {
        setSensitiveRequestId(res.data.data.request_id);
      }
      addLog(actionName, 'POST', path, payload, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'POST', path, payload, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, sensitiveRequest: false }));
    }
  };

  // 6. Review Sensitive Request Handler
  const handleReviewSensitiveRequest = async (e: React.FormEvent) => {
    e.preventDefault();
    const productId = parseInt(sensitiveProductId, 10);
    const actionName = 'SENSITIVE_LOGS.REQUEST.REVIEW';
    const path = `/api/v1/products/${productId}/sensitive-logs/requests/${sensitiveRequestId}/review`;
    setLoading((prev) => ({ ...prev, sensitiveReview: true }));

    const payload = {
      approval_status: sensitiveApprovalStatus,
    };

    try {
      const res = await getClient(true).post(path, payload);
      addLog(actionName, 'POST', path, payload, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'POST', path, payload, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, sensitiveReview: false }));
    }
  };

  // 7. Reveal Sensitive Value Handler
  const handleRevealSensitive = async (e: React.FormEvent) => {
    e.preventDefault();
    const productId = parseInt(sensitiveProductId, 10);
    const actionName = 'SENSITIVE_LOGS.REVEAL';
    const path = `/api/v1/products/${productId}/sensitive-logs/reveal`;
    setLoading((prev) => ({ ...prev, sensitiveReveal: true }));

    const payload = {
      request_id: sensitiveRequestId,
      password: sensitiveRevealPassword,
    };

    try {
      const res = await getClient(true).post(path, payload);
      addLog(actionName, 'POST', path, payload, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'POST', path, payload, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, sensitiveReveal: false }));
    }
  };

  // 8. List Sensitive Requests Handler
  const handleListSensitiveRequests = async (e: React.FormEvent) => {
    e.preventDefault();
    const productId = parseInt(sensitiveProductId, 10);
    const actionName = 'SENSITIVE_LOGS.REQUEST.LIST';
    const path = `/api/v1/products/${productId}/sensitive-logs/requests`;
    setLoading((prev) => ({ ...prev, sensitiveList: true }));

    try {
      const res = await getClient(true).get(path);
      addLog(actionName, 'GET', path, null, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'GET', path, null, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, sensitiveList: false }));
    }
  };

  // 9. List Sensitive History Handler
  const handleListSensitiveHistory = async (e: React.FormEvent) => {
    e.preventDefault();
    const productId = parseInt(sensitiveProductId, 10);
    const actionName = 'SENSITIVE_LOGS.HISTORY.LIST';
    const path = `/api/v1/products/${productId}/sensitive-logs/history`;
    setLoading((prev) => ({ ...prev, sensitiveHistory: true }));

    try {
      const res = await getClient(true).get(path);
      addLog(actionName, 'GET', path, null, res.status, res.data);
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      addLog(actionName, 'GET', path, null, status, data, err.message);
    } finally {
      setLoading((prev) => ({ ...prev, sensitiveHistory: false }));
    }
  };

  const handleClearAuth = () => {
    clearAuth();
    setManualToken('');
    addLog('CLEAR_SESSION', 'LOCAL', '-', null, 200, { message: 'Token cleared' });
  };

  if (view === 'catalog') {
    return (
      <div className="p-6">
        <div className="mb-4 flex gap-2">
          <button
            type="button"
            onClick={() => setView('console')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            API Console
          </button>
          <button
            type="button"
            onClick={() => setView('catalog')}
            className="border border-gray-900 px-3 py-1 text-sm bg-gray-900 text-white"
          >
            Product Catalog
          </button>
          <button
            type="button"
            onClick={() => setView('grant')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Grant Access
          </button>
          <button
            type="button"
            onClick={() => setView('roles')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Product Roles
          </button>
          <button
            type="button"
            onClick={() => setView('sensitive')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Sensitive Logs Console
          </button>
        </div>
        <ProductCatalogPage />
      </div>
    );
  }

  if (view === 'grant') {
    return (
      <div className="p-6">
        <div className="mb-4 flex gap-2">
          <button
            type="button"
            onClick={() => setView('console')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            API Console
          </button>
          <button
            type="button"
            onClick={() => setView('catalog')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Product Catalog
          </button>
          <button
            type="button"
            onClick={() => setView('grant')}
            className="border border-gray-900 px-3 py-1 text-sm bg-gray-900 text-white"
          >
            Grant Access
          </button>
          <button
            type="button"
            onClick={() => setView('roles')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Product Roles
          </button>
          <button
            type="button"
            onClick={() => setView('sensitive')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Sensitive Logs Console
          </button>
        </div>
        <AccessControlPage />
      </div>
    );
  }

  if (view === 'roles') {
    return (
      <div className="p-6">
        <div className="mb-4 flex gap-2">
          <button
            type="button"
            onClick={() => setView('console')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            API Console
          </button>
          <button
            type="button"
            onClick={() => setView('catalog')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Product Catalog
          </button>
          <button
            type="button"
            onClick={() => setView('grant')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Grant Access
          </button>
          <button
            type="button"
            onClick={() => setView('roles')}
            className="border border-gray-900 px-3 py-1 text-sm bg-gray-900 text-white"
          >
            Product Roles
          </button>
          <button
            type="button"
            onClick={() => setView('sensitive')}
            className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
          >
            Sensitive Logs Console
          </button>
        </div>
        <ProductRoleConsolePage />
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto font-sans bg-gray-50 min-h-screen text-gray-800">
      <div className="mb-4 flex gap-2">
        <button
          type="button"
          onClick={() => setView('console')}
          className={`border px-3 py-1 text-sm ${view === 'console' ? 'border-gray-900 bg-gray-900 text-white' : 'border-gray-300 bg-white hover:bg-gray-50'}`}
        >
          API Console
        </button>
        <button
          type="button"
          onClick={() => setView('catalog')}
          className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
        >
          Product Catalog
        </button>
        <button
          type="button"
          onClick={() => setView('grant')}
          className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
        >
          Grant Access
        </button>
        <button
          type="button"
          onClick={() => setView('roles')}
          className="border border-gray-300 px-3 py-1 text-sm bg-white hover:bg-gray-50"
        >
          Product Roles
        </button>
        <button
          type="button"
          onClick={() => setView('sensitive')}
          className={`border px-3 py-1 text-sm ${view === 'sensitive' ? 'border-gray-900 bg-gray-900 text-white' : 'border-gray-300 bg-white hover:bg-gray-50'}`}
        >
          Sensitive Logs Console
        </button>
      </div>
      <div className="mb-6 border-b border-gray-200 pb-4">
        <h1 className="text-3xl font-bold text-gray-900">
          {view === 'sensitive' ? 'Sensitive Logs & Approvals Console' : 'API Audit Log Test Console'}
        </h1>
        <p className="text-gray-600 mt-1">
          {view === 'sensitive'
            ? 'ใช้สำหรับทดสอบการทำงานของ Masked Data (Log Masking), การขอสิทธิ์ และการอนุมัติเข้าดูข้อมูลละเอียด'
            : 'ใช้ส่งคำขอทดสอบเพื่อดูการบันทึก Audit Logs ในหลังบ้าน (Vite + React + Tailwind)'}
        </p>
      </div>

      {/* Settings section */}
      <div className="bg-white p-4 rounded-lg shadow-sm border border-gray-200 mb-6 grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-semibold text-gray-700 mb-1">Backend API Base URL</label>
          <input
            type="text"
            value={baseURL}
            onChange={(e) => setBaseURL(e.target.value)}
            className="w-full p-2 border border-gray-300 rounded focus:ring-2 focus:ring-indigo-500 focus:outline-none"
          />
        </div>
        <div>
          <label className="block text-sm font-semibold text-gray-700 mb-1">
            Authorization Token (Bearer)
            {accessToken && <span className="text-xs text-green-600 ml-2 font-normal">(ดึงมาจาก Store อัตโนมัติ)</span>}
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="ไม่ต้องใส่หากต้องการทดสอบแบบ Unauthorized"
              value={manualToken}
              onChange={(e) => setManualToken(e.target.value)}
              className="flex-1 p-2 border border-gray-300 rounded text-xs focus:ring-2 focus:ring-indigo-500 focus:outline-none"
            />
            {manualToken && (
              <button
                type="button"
                onClick={handleClearAuth}
                className="bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600 transition"
              >
                Clear
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Main Grid: Forms (Left), Log View (Right) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        
        {/* Left Column: API Action Forms */}
        <div className="space-y-6">
          
          {view !== 'sensitive' ? (
            <>
              {/* Action 1: Register */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">1. Register User (Public API)</h2>
                <form onSubmit={handleRegister} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Email</label>
                      <input
                        type="email"
                        required
                        value={registerForm.email}
                        onChange={(e) => setRegisterForm({ ...registerForm, email: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Username</label>
                      <input
                        type="text"
                        required
                        value={registerForm.username}
                        onChange={(e) => setRegisterForm({ ...registerForm, username: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">First Name / Last Name</label>
                      <div className="flex gap-1">
                        <input
                          type="text"
                          placeholder="First"
                          required
                          value={registerForm.first_name}
                          onChange={(e) => setRegisterForm({ ...registerForm, first_name: e.target.value })}
                          className="w-1/2 p-1.5 border border-gray-300 rounded text-sm"
                        />
                        <input
                          type="text"
                          placeholder="Last"
                          required
                          value={registerForm.last_name}
                          onChange={(e) => setRegisterForm({ ...registerForm, last_name: e.target.value })}
                          className="w-1/2 p-1.5 border border-gray-300 rounded text-sm"
                        />
                      </div>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Password</label>
                      <input
                        type="text"
                        required
                        value={registerForm.password}
                        onChange={(e) => setRegisterForm({ ...registerForm, password: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Confirm Password</label>
                      <input
                        type="text"
                        required
                        value={registerForm.confirm_password}
                        onChange={(e) => setRegisterForm({ ...registerForm, confirm_password: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.register}
                    className="w-full bg-indigo-600 text-white p-2 rounded text-sm font-semibold hover:bg-indigo-700 disabled:opacity-50 transition"
                  >
                    {loading.register ? 'กำลังส่ง...' : 'ส่งคำขอ REGISTER (AUTH.REGISTER.CREATE)'}
                  </button>
                </form>
              </div>

              {/* Action 2: Login */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">2. Login User (Public API)</h2>
                <form onSubmit={handleLogin} className="space-y-3">
                  <div>
                    <label className="block text-xs font-semibold text-gray-600">Email Identifier</label>
                    <input
                      type="text"
                      required
                      placeholder="เช่น godmode@godmail.com หรืออีเมลที่เพิ่งสมัคร"
                      value={loginForm.identifier}
                      onChange={(e) => setLoginForm({ ...loginForm, identifier: e.target.value })}
                      className="w-full p-1.5 border border-gray-300 rounded text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-gray-600">Password</label>
                    <input
                      type="password"
                      required
                      value={loginForm.password}
                      onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })}
                      className="w-full p-1.5 border border-gray-300 rounded text-sm"
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={loading.login}
                    className="w-full bg-indigo-600 text-white p-2 rounded text-sm font-semibold hover:bg-indigo-700 disabled:opacity-50 transition"
                  >
                    {loading.login ? 'กำลังส่ง...' : 'ส่งคำขอ LOGIN (AUTH.LOGIN.CREATE)'}
                  </button>
                </form>
              </div>

              {/* Action 3: Give Admin Access */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">3. Give Admin Access (Admin Protected)</h2>
                <form onSubmit={handleGiveAdmin} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">User ID</label>
                      <input
                        type="number"
                        required
                        placeholder="ID ของผู้ใช้ที่สมัครเสร็จ"
                        value={giveAdminForm.id}
                        onChange={(e) => setGiveAdminForm({ ...giveAdminForm, id: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Role ID</label>
                      <select
                        value={giveAdminForm.role_id}
                        onChange={(e) => setGiveAdminForm({ ...giveAdminForm, role_id: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      >
                        <option value="1">1 (GOD Mode)</option>
                        <option value="2">2 (Owner)</option>
                        <option value="3">3 (Superadmin)</option>
                        <option value="4">4 (User)</option>
                      </select>
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.giveAdmin}
                    className="w-full bg-emerald-600 text-white p-2 rounded text-sm font-semibold hover:bg-emerald-700 disabled:opacity-50 transition"
                  >
                    {loading.giveAdmin ? 'กำลังส่ง...' : 'ส่งคำขอ GIVE ADMIN (ADMIN.GIVE_ADMIN.UPDATE)'}
                  </button>
                </form>
              </div>

              {/* Action 4: Create Product */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">4. Create Product (Auth Protected)</h2>
                <form onSubmit={handleCreateProduct} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Product Name</label>
                      <input
                        type="text"
                        required
                        value={productForm.product_name}
                        onChange={(e) => setProductForm({ ...productForm, product_name: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Product Code</label>
                      <input
                        type="text"
                        required
                        value={productForm.product_code}
                        onChange={(e) => setProductForm({ ...productForm, product_code: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Environment Code</label>
                      <input
                        type="text"
                        required
                        value={productForm.env_code}
                        onChange={(e) => setProductForm({ ...productForm, env_code: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Environment Name</label>
                      <input
                        type="text"
                        required
                        value={productForm.env_name}
                        onChange={(e) => setProductForm({ ...productForm, env_name: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.product}
                    className="w-full bg-emerald-600 text-white p-2 rounded text-sm font-semibold hover:bg-emerald-700 disabled:opacity-50 transition"
                  >
                    {loading.product ? 'กำลังส่ง...' : 'ส่งคำขอ CREATE PRODUCT (PRODUCTS.CREATE)'}
                  </button>
                </form>
              </div>

              {/* Action 5: Manual Log Ingestion */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">5. Ingest & Process Log (Manual Ingestion)</h2>
                <form onSubmit={handleIngestLog} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Product ID</label>
                      <input
                        type="number"
                        required
                        value={ingestForm.productId}
                        onChange={(e) => setIngestForm({ ...ingestForm, productId: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Environment ID</label>
                      <input
                        type="number"
                        required
                        value={ingestForm.environmentId}
                        onChange={(e) => setIngestForm({ ...ingestForm, environmentId: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Project ID (Optional)</label>
                      <input
                        type="number"
                        placeholder="ไม่บังคับ"
                        value={ingestForm.projectId}
                        onChange={(e) => setIngestForm({ ...ingestForm, projectId: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Category / Feature ID (Optional)</label>
                      <input
                        type="number"
                        placeholder="ไม่บังคับ"
                        value={ingestForm.categoryId}
                        onChange={(e) => setIngestForm({ ...ingestForm, categoryId: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Log Level</label>
                      <select
                        value={ingestForm.logLevel}
                        onChange={(e) => setIngestForm({ ...ingestForm, logLevel: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      >
                        <option value="INFO">INFO</option>
                        <option value="WARN">WARN</option>
                        <option value="ERROR">ERROR</option>
                        <option value="DEBUG">DEBUG</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Message</label>
                      <input
                        type="text"
                        required
                        value={ingestForm.message}
                        onChange={(e) => setIngestForm({ ...ingestForm, message: e.target.value })}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.ingest}
                    className="w-full bg-blue-600 text-white p-2 rounded text-sm font-semibold hover:bg-blue-700 disabled:opacity-50 transition"
                  >
                    {loading.ingest ? 'กำลังส่งและประมวลผล...' : 'ส่ง Log + trigger consume (Manual Process)'}
                  </button>
                </form>
              </div>
            </>
          ) : (
            <>
              {/* Action 5: Create Access Request */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">1. ขอสิทธิ์เข้าดูข้อมูลละเอียด (Create Access Request)</h2>
                <form onSubmit={handleCreateSensitiveRequest} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Product ID</label>
                      <input
                        type="number"
                        required
                        value={sensitiveProductId}
                        onChange={(e) => setSensitiveProductId(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Log ID (Elastic Document ID)</label>
                      <input
                        type="text"
                        placeholder="เช่น uuid ของ log"
                        value={sensitiveLogId}
                        onChange={(e) => setSensitiveLogId(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Field Path (เช่น payload.password)</label>
                      <input
                        type="text"
                        placeholder="เช่น password"
                        value={sensitiveFieldPath}
                        onChange={(e) => setSensitiveFieldPath(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">เหตุผลขอเข้าดู (Reason)</label>
                      <input
                        type="text"
                        required
                        value={sensitiveReason}
                        onChange={(e) => setSensitiveReason(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.sensitiveRequest}
                    className="w-full bg-indigo-600 text-white p-2 rounded text-sm font-semibold hover:bg-indigo-700 disabled:opacity-50 transition"
                  >
                    {loading.sensitiveRequest ? 'กำลังส่ง...' : 'ส่งคำขอสร้าง Request (SENSITIVE_LOGS.REQUEST.CREATE)'}
                  </button>
                </form>
              </div>

              {/* Action 6: Review Access Request */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">2. อนุมัติ/ปฏิเสธคำขอ (Review Access Request - GOD/Owner/Superadmin Only)</h2>
                <form onSubmit={handleReviewSensitiveRequest} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Request ID (UUID)</label>
                      <input
                        type="text"
                        required
                        placeholder="ดึงจากข้อ 1 อัตโนมัติ"
                        value={sensitiveRequestId}
                        onChange={(e) => setSensitiveRequestId(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">สถานะการอนุมัติ (Status)</label>
                      <select
                        value={sensitiveApprovalStatus}
                        onChange={(e) => setSensitiveApprovalStatus(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      >
                        <option value="APPROVED">APPROVED (อนุมัติ)</option>
                        <option value="REJECTED">REJECTED (ปฏิเสธ)</option>
                      </select>
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.sensitiveReview}
                    className="w-full bg-emerald-600 text-white p-2 rounded text-sm font-semibold hover:bg-emerald-700 disabled:opacity-50 transition"
                  >
                    {loading.sensitiveReview ? 'กำลังส่ง...' : 'อนุมัติคำขอ (SENSITIVE_LOGS.REQUEST.REVIEW)'}
                  </button>
                </form>
              </div>

              {/* Action 7: Reveal Sensitive Value */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">3. ถอดรหัสดูข้อมูลดิบ (Reveal Plaintext Value - Password Confirmation)</h2>
                <form onSubmit={handleRevealSensitive} className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">Request ID (UUID)</label>
                      <input
                        type="text"
                        required
                        placeholder="คำขอที่ได้รับการอนุมัติแล้ว"
                        value={sensitiveRequestId}
                        onChange={(e) => setSensitiveRequestId(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-gray-600">ยืนยัน Password ของคุณ</label>
                      <input
                        type="password"
                        required
                        value={sensitiveRevealPassword}
                        onChange={(e) => setSensitiveRevealPassword(e.target.value)}
                        className="w-full p-1.5 border border-gray-300 rounded text-sm"
                      />
                    </div>
                  </div>
                  <button
                    type="submit"
                    disabled={loading.sensitiveReveal}
                    className="w-full bg-purple-600 text-white p-2 rounded text-sm font-semibold hover:bg-purple-700 disabled:opacity-50 transition"
                  >
                    {loading.sensitiveReveal ? 'กำลังส่ง...' : 'ดึงค่าและถอดรหัส (SENSITIVE_LOGS.REVEAL)'}
                  </button>
                </form>
              </div>

              {/* Action 8: Utilities / Queries */}
              <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200">
                <h2 className="text-lg font-bold text-gray-900 border-b border-gray-100 pb-2 mb-3">4. ดึงข้อมูลประวัติการขอและดูประวัติการเข้าถึง (Auditing Tools)</h2>
                <div className="grid grid-cols-2 gap-3">
                  <button
                    onClick={handleListSensitiveRequests}
                    disabled={loading.sensitiveList}
                    className="bg-gray-800 text-white p-2.5 rounded text-xs font-semibold hover:bg-gray-900 disabled:opacity-50 transition"
                  >
                    {loading.sensitiveList ? 'กำลังโหลด...' : 'ดึงรายการคำขอทั้งหมด (GET REQUESTS)'}
                  </button>
                  <button
                    onClick={handleListSensitiveHistory}
                    disabled={loading.sensitiveHistory}
                    className="bg-gray-800 text-white p-2.5 rounded text-xs font-semibold hover:bg-gray-900 disabled:opacity-50 transition"
                  >
                    {loading.sensitiveHistory ? 'กำลังโหลด...' : 'ดูประวัติ Audit Access (GET HISTORY)'}
                  </button>
                </div>
              </div>
            </>
          )}

        </div>

        {/* Right Column: Execution Logs / Response View */}
        <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-200 flex flex-col h-[700px]">
          <div className="flex justify-between items-center border-b border-gray-100 pb-2 mb-3">
            <h2 className="text-lg font-bold text-gray-900">Console Log / API Response</h2>
            <button
              onClick={() => setApiLogs([])}
              className="text-xs text-gray-500 hover:text-red-500 font-semibold"
            >
              Clear Logs
            </button>
          </div>

          <div className="flex-1 overflow-y-auto space-y-3 pr-2 text-xs">
            {apiLogs.length === 0 ? (
              <div className="text-center text-gray-400 py-12">
                ยังไม่มีข้อมูลการส่งคำขอทดสอบ กดปุ่มส่งคำขอฝั่งซ้ายได้เลยครับ
              </div>
            ) : (
              apiLogs.map((log, idx) => (
                <div key={idx} className="border border-gray-200 rounded p-3 bg-gray-50">
                  <div className="flex justify-between items-center mb-1">
                    <span className="font-bold text-gray-700">[{log.timestamp}] {log.action}</span>
                    <span className={`px-2 py-0.5 rounded font-mono font-bold ${
                      log.status && log.status < 300 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {log.method} {log.status || 'ERROR'}
                    </span>
                  </div>
                  <div className="text-gray-500 font-mono mb-2">Request URL: {log.url}</div>
                  
                  {log.payload && (
                    <div className="mb-2">
                      <div className="font-semibold text-gray-600 mb-0.5">Payload:</div>
                      <pre className="p-1.5 bg-gray-900 text-green-400 rounded overflow-x-auto text-[10px] max-h-24">
                        {JSON.stringify(log.payload, null, 2)}
                      </pre>
                    </div>
                  )}

                  {log.response && (
                    <div>
                      <div className="font-semibold text-gray-600 mb-0.5">Response Body:</div>
                      <pre className="p-1.5 bg-gray-900 text-yellow-300 rounded overflow-x-auto text-[10px] max-h-40">
                        {JSON.stringify(log.response, null, 2)}
                      </pre>
                    </div>
                  )}

                  {log.error && (
                    <div className="mt-2 text-red-600 font-semibold font-mono">
                      Network Error: {log.error}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
              
      </div>
    </div>
  );
}
