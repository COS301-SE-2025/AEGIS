import { useNavigate } from "react-router-dom";

export const TenantAdminDashboard = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-background text-foreground p-6">
      <div className="max-w-6xl mx-auto space-y-6">
        <h1 className="text-3xl font-bold">Tenant Admin Dashboard</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl shadow bg-card">
            <h2 className="text-xl font-semibold mb-2">Team Management</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Register DFIR Managers and assign them to teams.
            </p>
            <button
              onClick={() => navigate("/team-registration")}
              className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded"
            >
              Register DFIR Admin
            </button>
          </div>

          <div className="p-6 rounded-2xl shadow bg-card">
            <h2 className="text-xl font-semibold mb-2">Team Overview</h2>
            <p className="text-sm text-muted-foreground mb-4">
              View DFIR teams and their status across the organization.
            </p>
            <button
              onClick={() => navigate("/teams")}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
            >
              View Teams
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
