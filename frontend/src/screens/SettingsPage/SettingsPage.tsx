import { useState } from "react";
import { LogOut, Trash2, Settings, UserCog, Shield, Bell } from "lucide-react";
import { Link } from "react-router-dom";

export const  SettingsPage = () => {
  const isAdmin = true; // Replace with actual admin check from auth context
  const [users, setUsers] = useState([
    { id: 1, name: "Analyst One", email: "analyst1@aegis.com" },
    { id: 2, name: "Responder Alpha", email: "responder@aegis.com" },
    { id: 3, name: "Manager Zeta", email: "manager@aegis.com" },
  ]);

  const handleRemoveUser = (userId: number) => {
    if (!isAdmin) return;
    setUsers(prev => prev.filter(user => user.id !== userId));
  };

  return (
    <div className="bg-black min-h-screen text-white px-8 py-10">
      <h1 className="text-3xl font-bold mb-8 flex items-center gap-2">
        <Settings className="w-6 h-6" /> Settings
      </h1>

      {/* Profile Settings */}
      <div className="bg-gray-900 rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Profile Settings</h2>
        <p className="text-gray-400 mb-4">Manage your AEGIS account preferences.</p>
        <Link to="/profile"><button className="bg-blue-600 px-4 py-2 rounded-lg hover:bg-blue-500 transition-colors">
          Edit Profile
        </button></Link>
      </div>

      {/* DFIR Settings */}
      <div className="bg-gray-900 rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5" /> DFIR Configuration
        </h2>
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <span className="text-gray-300">Alert Threshold Level</span>
            <select className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-1 text-white">
              <option>Low</option>
              <option>Moderate</option>
              <option>High</option>
              <option>Critical</option>
            </select>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-gray-300">Evidence Retention (Days)</span>
            <input
              type="number"
              defaultValue={90}
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-1 w-24 text-white"
            />
          </div>
          <div className="flex justify-between items-center">
            <span className="text-gray-300">Notification Preferences</span>
            <button className="flex items-center gap-2 text-blue-500 hover:text-blue-400">
              <Bell className="w-5 h-5" />
              Configure Alerts
            </button>
          </div>
        </div>
      </div>

      {/* User Management (Admin Only) */}
      {isAdmin && (
        <div className="bg-gray-900 rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
            <UserCog className="w-5 h-5" /> User Management
          </h2>
          <ul className="space-y-4">
            {users.map(user => (
              <li key={user.id} className="flex justify-between items-center bg-gray-800 p-3 rounded-lg">
                <div>
                  <p className="font-semibold">{user.name}</p>
                  <p className="text-gray-400 text-sm">{user.email}</p>
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

      {/* Logout */}
      <div className="bg-gray-900 rounded-lg p-6">
        <h2 className="text-xl font-semibold mb-4">Logout</h2>
        <Link
          to="/login"
          className="inline-flex items-center gap-2 bg-red-600 px-4 py-2 rounded-lg hover:bg-red-500 transition-colors"
        >
          <LogOut className="w-5 h-5" />
          Logout
        </Link>
      </div>
    </div>
  );
};

