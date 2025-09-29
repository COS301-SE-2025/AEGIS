import { useState } from "react";
import { useTheme } from "../../context/ThemeContext";


import {
  LogOut,
  Trash2,
  Settings,
  UserCog,
  Shield,
  Bell,
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
export const SettingsPage = () => {

  const isAdmin = true;
  const [users, setUsers] = useState([
    { id: 1, name: "Analyst One", email: "analyst1@aegis.com" },
    { id: 2, name: "Responder Alpha", email: "responder@aegis.com" },
    { id: 3, name: "Manager Zeta", email: "manager@aegis.com" },
  ]);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmNewPassword, setConfirmNewPassword] = useState('');


  const handleRemoveUser = (userId: number) => {
    if (!isAdmin) return;
    setUsers((prev) => prev.filter((user) => user.id !== userId));
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

      {/* DFIR Settings */}
      <div className="bg-card text-card-foreground rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5" /> DFIR Configuration
        </h2>
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <span className="text-muted-foreground">Alert Threshold Level</span>
            <select className="bg-input border border-border rounded-lg px-3 py-1 text-foreground">
              <option>Low</option>
              <option>Moderate</option>
              <option>High</option>
              <option>Critical</option>
            </select>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-muted-foreground">Evidence Retention (Days)</span>
            <input
              type="number"
              defaultValue={90}
              className="bg-input border border-border rounded-lg px-3 py-1 w-24 text-foreground"
            />
          </div>
          <div className="flex justify-between items-center">
            <span className="text-muted-foreground">Notification Preferences</span>
            <button className="flex items-center gap-2 text-primary hover:text-primary/80">
              <Bell className="w-5 h-5 text-primary" />
              Configure Alerts
            </button>
          </div>
        </div>
      </div>

      {/* User Management */}
      {isAdmin && (
        <div className="bg-card text-card-foreground rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
            <UserCog className="w-5 h-5" /> User Management
          </h2>
          <ul className="space-y-4">
            {users.map((user) => (
              <li
                key={user.id}
                className="flex justify-between items-center bg-muted p-3 rounded-lg"
              >
                <div>
                  <p className="font-semibold">{user.name}</p>
                  <p className="text-muted-foreground text-sm">{user.email}</p>
                </div>
                <button
                  onClick={() => handleRemoveUser(user.id)}
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
      {/* Register User */}
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
};
