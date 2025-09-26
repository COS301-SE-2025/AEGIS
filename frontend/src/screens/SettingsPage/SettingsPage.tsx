import { useState, useEffect } from "react";
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


import {
  LogOut,
  Trash2,
  Settings,
  UserCog,
  Shield,
  User,
} from "lucide-react";
import { Link } from "react-router-dom";
import { ColorPalettePicker } from "../../components/ui/ColorPalettePicker";

// Section for toggling and displaying theme customization
const ThemeCustomizationSection = () => {
  const { theme } = useTheme();
  const [customizing, setCustomizing] = useState(false);
  const themeLabels: Record<string, string> = {
    default: 'Default (Cool Greys)',
    light: 'Light (Warmer Blues)',
    dark: 'Dark',
  };
  return (
    <div className="mt-6">
      <label className="flex items-center gap-2 mb-4">
        <input
          type="checkbox"
          checked={customizing}
          onChange={e => setCustomizing(e.target.checked)}
          className="accent-primary"
        />
        <span className="font-medium">Customize current theme palette</span>
      </label>
      {customizing && (
        <>
          <div className="mb-2 text-muted-foreground text-sm">
            Customizing: <span className="font-semibold">{themeLabels[theme]}</span>
          </div>
          <ColorPalettePicker />
        </>
      )}
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

  useEffect(() => {
    // Check role and tenantId after mount (when sessionStorage is available)
    try {
      const token = sessionStorage.getItem('authToken');
      if (token) {
        const payload = JSON.parse(atob(token.split('.')[1]));
        setIsDFIRAdmin(payload.role === 'DFIR Admin');
        setTenantId(payload.tenant_id || payload.tenantId || null);
      }
    } catch {}
  }, []);

  useEffect(() => {
    if (isDFIRAdmin && tenantId) {
      axios.get(`http://localhost:8080/api/v1/tenants/${tenantId}/users`, {
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
          setUsers(data || []);
          setTotalUsers(total);
        })
        .catch(() => { setUsers([]); setTotalUsers(null); });
    }
  }, [isDFIRAdmin, tenantId, page, pageSize]);

  const handleRemoveUser = async (userId: string) => {
    if (!isDFIRAdmin) return;
    // Confirm with a toast instead of window.confirm
    const confirmed = await new Promise<boolean>((resolve) => {
      const confirmToast = document.createElement('div');
      confirmToast.style.position = 'fixed';
      confirmToast.style.bottom = '32px';
      confirmToast.style.left = '50%';
      confirmToast.style.transform = 'translateX(-50%)';
      confirmToast.style.background = '#334155';
      confirmToast.style.color = 'white';
      confirmToast.style.padding = '16px 32px';
      confirmToast.style.borderRadius = '8px';
      confirmToast.style.fontSize = '1rem';
      confirmToast.style.zIndex = '9999';
      confirmToast.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
      confirmToast.innerHTML = `Are you sure you want to remove this user?
        <button id="yesBtn" style="margin-left:16px;padding:4px 12px;background:#ef4444;color:white;border:none;border-radius:4px;cursor:pointer;">Yes</button>
        <button id="noBtn" style="margin-left:8px;padding:4px 12px;background:#64748b;color:white;border:none;border-radius:4px;cursor:pointer;">No</button>`;
      document.body.appendChild(confirmToast);
      confirmToast.querySelector('#yesBtn')?.addEventListener('click', () => { confirmToast.remove(); resolve(true); });
      confirmToast.querySelector('#noBtn')?.addEventListener('click', () => { confirmToast.remove(); resolve(false); });
    });
    if (!confirmed) return;
    try {
      await axios.delete(`http://localhost:8080/api/v1/users/${userId}`, {
        headers: { Authorization: `Bearer ${sessionStorage.getItem('authToken')}` },
      });
      setUsers((prev) => prev.filter((user) => (user.id || user.ID) !== userId));
      setTotalUsers((old) => (old !== null ? Math.max(0, old - 1) : null));
      showToast('User removed successfully.', 'success');
    } catch (err: any) {
      showToast(
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        'Failed to remove user.',
        'error'
      );
    }
  };

  return (
    <div className="min-h-screen px-8 py-10 bg-background text-foreground transition-colors">
     
      <div className="flex items-center justify-between border-b border-border pb-4 mb-6">
        {/* Left: Page title */}
        <h1 className="text-3xl font-bold flex items-center gap-2 text-foreground">
          <Settings className="w-6 h-6" /> Settings
        </h1>

        {/* Right: Navigation buttons */}
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

      {/* User Management */}
      {isDFIRAdmin && (
        <div className="bg-card text-card-foreground rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
            <UserCog className="w-5 h-5" /> User Management
          </h2>
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
                </div>
                <button
                  onClick={() => handleRemoveUser(user.id || user.ID)}
                  className="text-red-400 hover:text-red-300 flex items-center gap-1"
                >
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
      </div>

      {/* Logout */}
      <div className="bg-card text-card-foreground rounded-lg p-6">
        <h2 className="text-xl font-semibold mb-4">Logout</h2>
        <Link
          to="/login"
          className="inline-flex items-center gap-2 bg-destructive text-destructive-foreground px-4 py-2 rounded-lg hover:bg-destructive/80 transition-colors"
        >
          <LogOut className="w-5 h-5 text-destructive-foreground" />
          Logout
        </Link>
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
      <h2 className="text-2xl font-bold mb-4">Customize Theme</h2>
      <ThemeToggle />
      <ThemeCustomizationSection />
    </div>
      {showPasswordModal && (
      <div className="fixed inset-0 bg-background flex items-center justify-center z-50">
        <div className="bg-card text-card-foreground rounded-lg p-6 w-full max-w-md shadow-lg">
          <h2 className="text-xl font-semibold mb-4">Reset Password</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-muted-foreground mb-1">Old Password</label>
              <input
                type="password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm text-muted-foreground mb-1">New Password</label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm text-muted-foreground mb-1">Confirm New Password</label>
              <input
                type="password"
                value={confirmNewPassword}
                onChange={(e) => setConfirmNewPassword(e.target.value)}
                className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none"
              />
            </div>
          </div>
          <div className="mt-6 flex justify-end gap-2">
            <button
              onClick={() => setShowPasswordModal(false)}
              className="px-4 py-2 rounded-lg bg-muted hover:bg-muted/70 text-muted-foreground"
            >
              Cancel
            </button>
            <button
              onClick={() => {
                // TODO: Implement reset logic
                setShowPasswordModal(false);
              }}
              className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    )}

    </div>
  );
}

export { SettingsPage };
