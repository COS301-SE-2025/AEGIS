import { useNavigate } from "react-router-dom";

export const SystemAdminDashboard = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-background text-foreground p-6">
      <div className="max-w-6xl mx-auto space-y-6">
        <h1 className="text-3xl font-bold">System Admin Dashboard</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl shadow bg-card">
            <h2 className="text-xl font-semibold mb-2">Tenant Management</h2>
            <p className="text-sm text-muted-foreground mb-4">
              View and manage all registered tenants on the system.
            </p>
            <button
              onClick={() => navigate("/tenants")}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
            >
              View Tenants
            </button>
          </div>

          <div className="p-6 rounded-2xl shadow bg-card">
            <h2 className="text-xl font-semibold mb-2">Register Tenant Admin</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Create new tenant admins by registering their organization.
            </p>
            <button
              onClick={() => navigate("/tenant-registration")}
              className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded"
            >
              Register Tenant Admin
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
