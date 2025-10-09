import { useState, useEffect } from "react";
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
// Simple toast implementation (replace with your UI lib if available)
function showToast(message: string, type: 'success' | 'error' = 'success') {
  const toast = document.createElement('div');
  toast.textContent = message;
  toast.style.position = 'fixed';
  toast.style.bottom = '32px';
  toast.style.left = '50%';
  toast.style.transform = 'translateX(-50%)';
  toast.style.background = type === 'success' ? '#22c55e' : '#ef4444';
  toast.style.color = 'white';
  toast.style.padding = '12px 24px';
  toast.style.borderRadius = '8px';
  toast.style.fontSize = '1rem';
  toast.style.zIndex = '9999';
  toast.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
  document.body.appendChild(toast);
  setTimeout(() => { toast.remove(); }, 2500);
}
import axios from "axios";
import { useTheme } from "../../context/ThemeContext";
import { ArrowLeft } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";

import {
  LogOut,
  Trash2,
  Settings,
  UserCog,
  Shield,
  User,
  Lock,
  Eye,
  EyeOff,
  AlertTriangle,
  X,
} from "lucide-react";

// MFA Authentication Modal Component
const MFAAuthModal = ({ 
  isOpen, 
  onClose, 
  onConfirm, 
  userToRemove,
  loading 
}: {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (password: string, mfaCode?: string) => void;
  userToRemove: any;
  loading: boolean;
}) => {
  const [password, setPassword] = useState('');
  const [mfaCode, setMfaCode] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [step, setStep] = useState<'password' | 'mfa'>('password');
  const [error, setError] = useState('');


const handlePasswordSubmit = async () => {
  if (!password.trim()) {
    setError('Password is required');
    return;
  }

  try {
    const token = sessionStorage.getItem('authToken');
    
    if (!token) {
      setError('Authentication token not found. Please log in again.');
      return;
    }

    console.log('Sending verify-admin request...');

    const response = await axios.post('https://localhost/api/v1/auth/verify-admin', {
      password: password
    }, {
      headers: { 
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    const data = response.data as { valid: boolean; message?: string };
    
    if (data.valid) {
      // Password is valid, proceed with user removal
      onConfirm(password);
      setError('');
    } else {
      setError(data.message || 'Invalid password');
    }
  } catch (err: any) {
    console.error('Admin verification error:', err);
    
    if (err.response?.status === 401) {
      setError('Authentication failed. Please log in again.');
    } else if (err.response?.status === 403) {
      setError('Admin access required.');
    } else {
      setError(err?.response?.data?.error || err?.response?.data?.message || 'Verification failed');
    }
  }
};

  const handleMFASubmit = () => {
    if (!mfaCode.trim()) {
      setError('MFA code is required');
      return;
    }
    onConfirm(password, mfaCode);
  };

  const handleClose = () => {
    setPassword('');
    setMfaCode('');
    setStep('password');
    setError('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-card text-card-foreground rounded-lg p-6 w-full max-w-md shadow-xl border border-border">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-destructive" />
            <h2 className="text-xl font-semibold">Admin Authentication Required</h2>
          </div>
          <button
            onClick={handleClose}
            className="text-muted-foreground hover:text-foreground"
            disabled={loading}
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
          <p className="text-sm text-destructive font-medium">
            You are about to remove:
          </p>
          <p className="text-sm text-foreground mt-1">
            <strong>{userToRemove?.full_name || userToRemove?.FullName || userToRemove?.name}</strong>
          </p>
          <p className="text-xs text-muted-foreground">
            {userToRemove?.email || userToRemove?.Email}
          </p>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
            <p className="text-sm text-destructive">{error}</p>
          </div>
        )}

        {step === 'password' && (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Please enter your admin password to continue:
            </p>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                placeholder="Enter your password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  setError('');
                }}
                className="w-full px-3 py-2 pr-10 bg-input border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                disabled={loading}
                autoFocus
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                disabled={loading}
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            <div className="flex justify-end gap-2">
              <button
                onClick={handleClose}
                className="px-4 py-2 rounded-lg bg-muted hover:bg-muted/70 text-muted-foreground"
                disabled={loading}
              >
                Cancel
              </button>
              <button
                onClick={handlePasswordSubmit}
                className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
                disabled={loading || !password.trim()}
              >
                {loading ? 'Verifying...' : 'Continue'}
              </button>
            </div>
          </div>
        )}

        {step === 'mfa' && (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Please enter your MFA code to complete the authentication:
            </p>
            <div>
              <input
                type="text"
                placeholder="Enter 6-digit MFA code"
                value={mfaCode}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, '').slice(0, 6);
                  setMfaCode(value);
                  setError('');
                }}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary text-center text-lg tracking-widest"
                disabled={loading}
                autoFocus
                maxLength={6}
              />
            </div>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setStep('password')}
                className="px-4 py-2 rounded-lg bg-muted hover:bg-muted/70 text-muted-foreground"
                disabled={loading}
              >
                Back
              </button>
              <button
                onClick={handleMFASubmit}
                className="px-4 py-2 rounded-lg bg-destructive text-destructive-foreground hover:bg-destructive/90 disabled:opacity-50"
                disabled={loading || mfaCode.length !== 6}
              >
                {loading ? 'Removing User...' : 'Remove User'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};



// ThemeToggle component for switching between themes
const ThemeToggle = () => {
  const { theme, setTheme } = useTheme();
  return (
    <div className="flex items-center gap-4 mb-4">
      <label className="font-medium">Theme:</label>
      <select
        className="bg-input border border-border rounded-lg px-3 py-2 text-foreground focus:outline-none"
        value={theme}
        onChange={e => setTheme(e.target.value as any)}
      >
        <option value="default">Default</option>
        <option value="light">Neutral</option>
        <option value="dark">Dark</option>
      </select>
    </div>
  );
};

function SettingsPage() {
  const navigate = useNavigate();
  const [users, setUsers] = useState<any[]>([]);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmNewPassword, setConfirmNewPassword] = useState('');
  const [isDFIRAdmin, setIsDFIRAdmin] = useState(false);
  const [tenantId, setTenantId] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalUsers, setTotalUsers] = useState<number | null>(null);
  
  // Add state for user role
  const [userRole, setUserRole] = useState<string>('');
  
  // MFA Authentication State
  const [showMFAModal, setShowMFAModal] = useState(false);
  const [userToRemove, setUserToRemove] = useState<any>(null);
  const [removingUser, setRemovingUser] = useState(false);

  // Add state for password reset loading and errors
  const [passwordResetLoading, setPasswordResetLoading] = useState(false);
  const [passwordResetError, setPasswordResetError] = useState('');

  useEffect(() => {
    // Check role and tenantId after mount (when sessionStorage is available)
    try {
      const token = sessionStorage.getItem('authToken');
      if (token) {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const role = payload.role || '';
        setUserRole(role);
        setIsDFIRAdmin(role === 'DFIR Admin');
        setTenantId(payload.tenant_id || payload.tenantId || null);
      }
    } catch {}
  }, []);

  useEffect(() => {
    if (isDFIRAdmin && tenantId) {
      axios.get(`https://localhost/api/v1/tenants/${tenantId}/users`, {
        headers: { Authorization: `Bearer ${sessionStorage.getItem('authToken')}` },
        params: { page, page_size: pageSize }
      })
        .then(res => {
          let data = (res.data as any)?.data;
          let total = (res.data as any)?.total || null;
          if (typeof data === 'undefined' && Array.isArray((res.data as any)?.users)) {
            data = (res.data as any).users;
            total = (res.data as any).total || null;
          }
          // Exclude DFIR Admin and Tenant Admin from the list
          const filtered = (data || []).filter((u: any) => {
            const role = u.role || u.Role;
            return role !== 'DFIR Admin' && role !== 'Tenant Admin';
          });
          setUsers(filtered);
          setTotalUsers(total !== null ? filtered.length : total);
        })
        .catch(() => { setUsers([]); setTotalUsers(null); });
    }
  }, [isDFIRAdmin, tenantId, page, pageSize]);

  const handleRemoveUser = async (userId: string) => {
    if (!isDFIRAdmin) return;
    
    const user = users.find(u => (u.id || u.ID) === userId);
    if (!user) return;

    setUserToRemove(user);
    setShowMFAModal(true);
  };

  const handleMFAConfirm = async (password: string, mfaCode?: string) => {
    if (!userToRemove) return;
    
    setRemovingUser(true);
    
    try {
      const token = sessionStorage.getItem('authToken');
      const payload = {
        password: password,
        ...(mfaCode && { mfaCode: mfaCode })
      };

      await axios.delete(
        `https://localhost/api/v1/users/${userToRemove.id || userToRemove.ID}`,
        {
          params: payload, // Send authentication data as query params
          headers: { Authorization: `Bearer ${token}` }
        }
      );

      setUsers((prev) => prev.filter((user) => (user.id || user.ID) !== (userToRemove.id || userToRemove.ID)));
      setTotalUsers((old) => (old !== null ? Math.max(0, old - 1) : null));
      showToast('User removed successfully.', 'success');
      
      setShowMFAModal(false);
      setUserToRemove(null);
    } catch (err: any) {
      showToast(
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        'Failed to remove user.',
        'error'
      );
    } finally {
      setRemovingUser(false);
    }
  };

  const handleMFAClose = () => {
    setShowMFAModal(false);
    setUserToRemove(null);
  };

      const handleLogout = async () => {
        try {
          const token = sessionStorage.getItem('authToken');
          
          if (token) {
            // Call backend logout endpoint
            try {
              await axios.post('https://localhost/api/v1/auth/logout', {}, {
                headers: { 
                  Authorization: `Bearer ${token}`,
                  'Content-Type': 'application/json'
                }
              });
              console.log('Server logout successful');
            } catch (err) {
              console.warn('Server logout failed, proceeding with client logout:', err);
            }
          }

          // Clear client-side storage
          sessionStorage.removeItem('authToken');
          localStorage.removeItem('authToken');
          sessionStorage.clear();
          
          showToast('Logged out successfully', 'success');
          navigate('/login');
          
        } catch (error) {
          console.error('Logout error:', error);
          
          // Force logout even if there's an error
          sessionStorage.removeItem('authToken');
          localStorage.removeItem('authToken');
          sessionStorage.clear();
          
          showToast('Logged out', 'success');
          navigate('/login');
        }
      };
  // Determine if navbar should be hidden
  const shouldHideNavbar = userRole === 'Tenant Admin' || userRole === 'System Admin';

  // Handle back navigation
  const handleBack = () => {
    navigate(-1); // Go back to previous page
  };

  // Add password validation function
  const validatePassword = (password: string) => {
    if (password.length < 8) {
      return 'Password must be at least 8 characters long';
    }
    if (!/(?=.*[a-z])/.test(password)) {
      return 'Password must contain at least one lowercase letter';
    }
    if (!/(?=.*[A-Z])/.test(password)) {
      return 'Password must contain at least one uppercase letter';
    }
    if (!/(?=.*\d)/.test(password)) {
      return 'Password must contain at least one number';
    }
    return '';
  };

  // Add the handlePasswordReset function
  const handlePasswordReset = async () => {
    // Validation
    if (!oldPassword.trim()) {
      setPasswordResetError('Current password is required');
      return;
    }

    if (!newPassword.trim()) {
      setPasswordResetError('New password is required');
      return;
    }

    if (!confirmNewPassword.trim()) {
      setPasswordResetError('Please confirm your new password');
      return;
    }

    if (newPassword !== confirmNewPassword) {
      setPasswordResetError('New passwords do not match');
      return;
    }

    // Password strength validation
    const passwordError = validatePassword(newPassword);
    if (passwordError) {
      setPasswordResetError(passwordError);
      return;
    }

    setPasswordResetLoading(true);
    setPasswordResetError('');

    try {
      const token = sessionStorage.getItem('authToken');
      
      if (!token) {
        setPasswordResetError('Authentication token not found. Please log in again.');
        return;
      }

      await axios.post('https://localhost/api/v1/auth/change-password', {
        oldPassword: oldPassword,
        newPassword: newPassword,
        confirmPassword: confirmNewPassword
      }, {
        headers: { 
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      // Success
      showToast('Password changed successfully!', 'success');
      setShowPasswordModal(false);
      navigate('/login');
      
      // Clear form
      setOldPassword('');
      setNewPassword('');
      setConfirmNewPassword('');
      setPasswordResetError('');

    } catch (err: any) {
      console.error('Password reset error:', err);
      
      if (err.response?.status === 401) {
        setPasswordResetError('Authentication failed. Please log in again.');
      } else if (err.response?.status === 400) {
        setPasswordResetError(err.response.data?.error || 'Invalid password data');
      } else {
        setPasswordResetError(err?.response?.data?.error || 'Failed to change password');
      }
    } finally {
      setPasswordResetLoading(false);
    }
  };

  return (
    <div className="min-h-screen px-8 py-10 bg-background text-foreground transition-colors">
      <ToastContainer position="top-center" aria-label="Notification Toasts" />
     
      <div className="flex items-center justify-between border-b border-border pb-4 mb-6">
        {/* Left: Page title */}
        <h1 className="text-3xl font-bold flex items-center gap-2 text-foreground">
          <Settings className="w-6 h-6" /> Settings
        </h1>

        {/* Right: Navigation - Show navbar or back button based on user role */}
        {shouldHideNavbar ? (
          <button
            onClick={handleBack}
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors border border-border hover:bg-muted"
          >
            <ArrowLeft className="w-4 h-4" />
            Back
          </button>
        ) : (
          <div className="flex items-center gap-4">
            <Link to="/dashboard">
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Dashboard
              </button>
            </Link>
            <Link to="/case-management">
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Case Management
              </button>
            </Link>
            <Link to="/evidence-viewer">
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Evidence Viewer
              </button>
            </Link>
            <Link to="/secure-chat">
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Secure Chat
              </button>
            </Link>
          </div>
        )}
      </div>

      {/* Profile Settings */}
      <div className="bg-card text-card-foreground rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Profile Settings</h2>
        <p className="text-muted-foreground mb-4">
          Manage your AEGIS account preferences.
        </p>
        <Link to="/profile">
          <button className="bg-primary text-primary-foreground px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors">
            Edit Profile
          </button>
        </Link>
      </div>

      {/* User Management - Only show for DFIR Admin */}
      {isDFIRAdmin && (
        <div className="bg-card text-card-foreground rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
            <UserCog className="w-5 h-5" /> User Management
          </h2>
          <div className="mb-4 p-3 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <div className="flex items-center gap-2 mb-1">
              <Shield className="w-4 h-4 text-amber-500" />
              <span className="text-sm font-medium text-amber-600">Security Notice</span>
            </div>
            <p className="text-xs text-amber-700">
              Removing users requires admin authentication and MFA verification for security.
            </p>
          </div>
          <div className="flex items-center justify-between mb-2">
            <div>
              <span className="text-sm text-muted-foreground">Page {page}</span>
              {totalUsers !== null && (
                <span className="ml-2 text-sm text-muted-foreground">Total Users: {totalUsers}</span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <button
                className="px-2 py-1 rounded bg-muted text-foreground border border-border disabled:opacity-50"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Prev
              </button>
              <button
                className="px-2 py-1 rounded bg-muted text-foreground border border-border disabled:opacity-50"
                onClick={() => setPage((p) => p + 1)}
                disabled={users.length < pageSize}
              >
                Next
              </button>
              <select
                className="ml-2 px-2 py-1 rounded border border-border bg-muted text-foreground"
                value={pageSize}
                onChange={e => { setPageSize(Number(e.target.value)); setPage(1); }}
              >
                {[5, 10, 20, 50].map(size => (
                  <option key={size} value={size}>{size} per page</option>
                ))}
              </select>
            </div>
          </div>
          <ul className="space-y-4">
            {users.length === 0 && (
              <li className="text-muted-foreground">No users found for this tenant.</li>
            )}
            {users.map((user) => (
              <li
                key={user.id || user.ID}
                className="flex justify-between items-center bg-muted p-3 rounded-lg"
              >
                <div>
                  <p className="font-semibold">{user.full_name || user.FullName || user.name}</p>
                  <p className="text-muted-foreground text-sm">{user.email || user.Email}</p>
                  <p className="text-sm text-muted-foreground">Role: {user.role || user.Role || 'N/A'}</p>
                </div>
                <button
                  onClick={() => handleRemoveUser(user.id || user.ID)}
                  className="text-red-400 hover:text-red-300 flex items-center gap-1 px-3 py-2 rounded-lg border border-red-400/20 hover:bg-red-400/10 transition-colors"
                  disabled={removingUser}
                >
                  <Shield className="w-4 h-4" />
                  <Trash2 className="w-4 h-4" />
                  Remove
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="bg-card text-card-foreground rounded-lg p-6 mt-6">
        <h2 className="text-xl font-semibold mb-4">Reset Password</h2>
        <button
          onClick={() => setShowPasswordModal(true)}
          className="inline-flex items-center gap-2 bg-secondary text-secondary-foreground px-4 py-2 rounded-lg hover:bg-secondary/80 transition-colors"
        >
          <Shield className="w-5 h-5 text-secondary-foreground" />
          Reset Password
        </button>

      {/* Logout */}
      <h2 className="text-xl font-semibold mb-4">Logout</h2>
      <button
        onClick={handleLogout}
        className="inline-flex items-center gap-2 bg-destructive text-destructive-foreground px-4 py-2 rounded-lg hover:bg-destructive/80 transition-colors"
      >
        <LogOut className="w-5 h-5 text-destructive-foreground" />
        Logout
      </button>

      </div>

      {/* Register User (DFIR Admin only) */}
      {isDFIRAdmin && (
        <div className="bg-card text-card-foreground rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-4">Register User</h2>
          <Link
            to="/register"
            className="inline-flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors"
          >
            <User className="w-5 h-5 text-primary-foreground" />
            Register User
          </Link>
        </div>
      )}

    <div className="p-8">
      <ThemeToggle />
    </div>

    {/* MFA Authentication Modal */}
    <MFAAuthModal
      isOpen={showMFAModal}
      onClose={handleMFAClose}
      onConfirm={handleMFAConfirm}
      userToRemove={userToRemove}
      loading={removingUser}
    />

    {showPasswordModal && (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-card text-card-foreground rounded-lg p-6 w-full max-w-md shadow-xl border border-border">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold flex items-center gap-2">
              <Lock className="w-5 h-5" />
              Reset Password
            </h2>
            <button
              onClick={() => {
                setShowPasswordModal(false);
                setPasswordResetError('');
                setOldPassword('');
                setNewPassword('');
                setConfirmNewPassword('');
              }}
              className="text-muted-foreground hover:text-foreground"
              disabled={passwordResetLoading}
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {passwordResetError && (
            <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
              <p className="text-sm text-destructive">{passwordResetError}</p>
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label className="block text-sm text-muted-foreground mb-1">Current Password</label>
              <input
                type="password"
                value={oldPassword}
                onChange={(e) => {
                  setOldPassword(e.target.value);
                  setPasswordResetError('');
                }}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                disabled={passwordResetLoading}
                placeholder="Enter your current password"
              />
            </div>
            <div>
              <label className="block text-sm text-muted-foreground mb-1">New Password</label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => {
                  setNewPassword(e.target.value);
                  setPasswordResetError('');
                }}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                disabled={passwordResetLoading}
                placeholder="Enter new password (min 8 characters)"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Must contain uppercase, lowercase, number, and be at least 8 characters
              </p>
            </div>
            <div>
              <label className="block text-sm text-muted-foreground mb-1">Confirm New Password</label>
              <input
                type="password"
                value={confirmNewPassword}
                onChange={(e) => {
                  setConfirmNewPassword(e.target.value);
                  setPasswordResetError('');
                }}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                disabled={passwordResetLoading}
                placeholder="Confirm your new password"
              />
            </div>
          </div>
          <div className="mt-6 flex justify-end gap-2">
            <button
              onClick={() => {
                setShowPasswordModal(false);
                setPasswordResetError('');
                setOldPassword('');
                setNewPassword('');
                setConfirmNewPassword('');
              }}
              className="px-4 py-2 rounded-lg bg-muted hover:bg-muted/70 text-muted-foreground"
              disabled={passwordResetLoading}
            >
              Cancel
            </button>
            <button
              onClick={handlePasswordReset}
              className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
              disabled={passwordResetLoading || !oldPassword || !newPassword || !confirmNewPassword}
            >
              {passwordResetLoading ? 'Changing...' : 'Change Password'}
            </button>
          </div>
        </div>
      </div>
    )}

    </div>
  );
}

export { SettingsPage };
