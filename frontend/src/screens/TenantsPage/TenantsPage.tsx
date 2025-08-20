import React, { useState, useMemo, useEffect } from "react";
import { Modal } from "../../components/ui/Modal";

interface Tenant {
  id: string;
  name: string;
  createdAt: string;
  status?: "active" | "inactive" | "pending"; // Optional status field
  
}

const ITEMS_PER_PAGE = 5;

export const TenantsPage = () => {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);


  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [newTenantName, setNewTenantName] = useState("");

  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  const [removeModalOpen, setRemoveModalOpen] = useState(false);
  const [tenantToRemove, setTenantToRemove] = useState<Tenant | null>(null);

const filteredTenants = useMemo(() => {
  return tenants.filter((t) =>
    (t.name || "").toLowerCase().includes(searchTerm.toLowerCase())
  );
}, [tenants, searchTerm]);
const paginatedTenants = useMemo(() => {
  const start = (currentPage - 1) * ITEMS_PER_PAGE;
  const end = start + ITEMS_PER_PAGE;
  return filteredTenants.slice(start, end);
}, [filteredTenants, currentPage]);

const mockStatus = () => {
  const statuses = ["active", "inactive", "pending"];
  return statuses[Math.floor(Math.random() * statuses.length)];
};

  
useEffect(() => {
  const fetchTenants = async () => {
    try {
      const res = await fetch("http://localhost:8080/api/v1/tenants");
      const data = await res.json();

      const tenantsRaw = Array.isArray(data) ? data : data.tenants;

      if (tenantsRaw) {
      const tenantsWithStatus = tenantsRaw.map((tenant: any) => ({
      id: tenant.ID,
      name: tenant.Name,
      createdAt: tenant.CreatedAt,
      updatedAt: tenant.UpdatedAt,
      status: mockStatus(),
      }));
      setTenants(tenantsWithStatus);

        console.log("Tenants fetched:", tenantsWithStatus); // ✅ HERE
      }

    } catch (error) {
      console.error("Failed to fetch tenants:", error);
    } finally {
      setLoading(false);
    }
  };

  fetchTenants();
}, []); 
  if (loading) {
    return <div className="text-center text-muted-foreground">Loading tenants...</div>;
  }

  const totalPages = Math.ceil(filteredTenants.length / ITEMS_PER_PAGE);

  const handleAddTenant = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTenantName.trim()) return;

    const newTenant: Tenant = {
      id: String(Date.now()),
      name: newTenantName.trim(),
      createdAt: new Date().toISOString().split("T")[0],
    };

    setTenants((prev) => [...prev, newTenant]);
    setNewTenantName("");
    setIsAddModalOpen(false);
    setCurrentPage(totalPages); // jump to last page where new tenant likely is
  };

  const confirmRemoveTenant = (tenant: Tenant) => {
    setTenantToRemove(tenant);
    setRemoveModalOpen(true);
  };

  const handleRemoveTenant = () => {
    if (!tenantToRemove) return;
    setTenants((prev) => prev.filter((t) => t.id !== tenantToRemove.id));
    setRemoveModalOpen(false);
    setTenantToRemove(null);
    if ((filteredTenants.length - 1) <= (currentPage - 1) * ITEMS_PER_PAGE) {
      setCurrentPage(Math.max(1, currentPage - 1));
    }
  };

  const btnBase =
    "px-3 py-1.5 rounded-md font-medium text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition";

  return (
    <div className="min-h-screen bg-white dark:bg-background text-foreground p-6">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-semibold">Tenants</h1>
          
        </div>

        {/* Search */}
        <div className="mb-4">
          <input
            type="text"
            placeholder="Search tenants..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setCurrentPage(1);
            }}
            className="w-full max-w-md px-3 py-2 rounded-md border border-input bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* Tenants List */}
        <div className="space-y-4">
          {filteredTenants.length === 0 ? (
            <p className="text-muted-foreground text-center">
              No tenants found.
            </p>
          ) : (
              paginatedTenants.map((tenant) => {
              if (!tenant.name || !tenant.createdAt || !tenant.status) return null;

              return (
                <div key={tenant.id} className="flex justify-between items-center border rounded-xl p-4 shadow-md bg-card dark:bg-gray-900">
                  <div>
                    <p className="text-lg font-medium">{tenant.name}</p>
                    <p className="text-sm text-muted-foreground">Created: {tenant.createdAt ? new Date(tenant.createdAt).toLocaleString() :"Unknown"}</p>
                    <p className="text-sm text-muted-foreground">
                      Status:{" "}
                      <span
                        className={`font-semibold ${
                          tenant.status === "active"
                            ? "text-green-600"
                            : tenant.status === "inactive"
                            ? "text-gray-500"
                            : "text-yellow-600"
                        }`}
                      >
                        {tenant.status}
                      </span>
                    </p>
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={() => alert(`Viewing details for ${tenant.name}`)}
                      className={`${btnBase} border border-border bg-muted hover:bg-muted/80 text-foreground`}
                    >
                      View Details
                    </button>
                    <button
                      onClick={() => confirmRemoveTenant(tenant)}
                      className={`${btnBase} bg-red-600 text-white hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600`}
                    >
                      Remove
                    </button>
                  </div>
                </div>
              );
              })
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center items-center gap-4 mt-6 text-sm text-muted-foreground">
            <button
              disabled={currentPage === 1}
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              className={`${btnBase} border border-border disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              Previous
            </button>
            <span>
              Page {currentPage} of {totalPages}
            </span>
            <button
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              className={`${btnBase} border border-border disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              Next
            </button>
          </div>
        )}

        <Modal isOpen={isAddModalOpen} onClose={() => setIsAddModalOpen(false)}>
          <h2 className="text-xl font-semibold mb-4">Add a New Tenant</h2>
          <form onSubmit={handleAddTenant} className="space-y-4">
            <div>
              <label htmlFor="tenantName" className="block text-sm font-medium mb-1">
                Tenant Name
              </label>
              <input
                id="tenantName"
                type="text"
                value={newTenantName}
                onChange={(e) => setNewTenantName(e.target.value)}
                className="w-full px-3 py-2 rounded-md border border-input bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
                autoFocus
              />
            </div>
            <div className="flex justify-end gap-3">
              <button
                type="button"
                onClick={() => setIsAddModalOpen(false)}
                className={`${btnBase} border border-border bg-muted hover:bg-muted/70 text-foreground`}
              >
                Cancel
              </button>
              <button
                type="submit"
                className={`${btnBase} bg-blue-600 text-white hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600`}
              >
                Save
              </button>
            </div>
          </form>
        </Modal>

        {/* Remove Tenant Confirmation Modal */}
        <Modal isOpen={removeModalOpen} onClose={() => setRemoveModalOpen(false)}>
          <h2 className="text-xl font-semibold mb-4">Confirm Removal</h2>
          <p>
            Are you sure you want to remove tenant{" "}
            <strong>{tenantToRemove?.name}</strong>? This action cannot be undone.
          </p>
          <div className="flex justify-end gap-3 mt-6">
            <button
              onClick={() => setRemoveModalOpen(false)}
              className={`${btnBase} border border-border bg-muted hover:bg-muted/70 text-foreground`}
            >
              Cancel
            </button>
            <button
              onClick={handleRemoveTenant}
              className={`${btnBase} bg-red-600 text-white hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600`}
            >
              Remove
            </button>
          </div>
        </Modal>
      </div>
    </div>
  );
};
