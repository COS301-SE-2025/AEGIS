import React, { useState, useMemo, useEffect } from "react";
import { Modal } from "../../components/ui/Modal";
import axios from "axios";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router-dom";

interface Tenant {
  id: string;
  name: string;
  createdAt: string;
}
// Simple toast implementation
const Toast = ({ message, type, onClose }: { 
  message: string; 
  type: 'success' | 'error'; 
  onClose: () => void;
}) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, 5000); // Auto close after 5 seconds

    return () => clearTimeout(timer);
  }, [onClose]);

  return (
    <div className={`fixed top-4 right-4 z-50 p-4 rounded-md shadow-lg flex items-center gap-3 ${
      type === 'success' 
        ? 'bg-green-600 text-white' 
        : 'bg-red-600 text-white'
    }`}>
      <div className="flex-1">
        {message}
      </div>
      <button 
        onClick={onClose}
        className="text-white hover:text-gray-200 text-xl leading-none"
      >
        ×
      </button>
    </div>
  );
};

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
  const [verificationError, setVerificationError] = useState("");
  const [isRemoving, setIsRemoving] = useState(false);
const navigate = useNavigate();

  // Toast state
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  // Function to show toast
  const showToast = (message: string, type: 'success' | 'error') => {
    setToast({ message, type });
  };

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

  useEffect(() => {
    const fetchTenants = async () => {
      try {
        const res = await fetch("https://localhost/api/v1/tenants");
        const data = await res.json();

        const tenantsRaw = Array.isArray(data) ? data : data.tenants;

        if (tenantsRaw) {
          const tenantsWithoutStatus = tenantsRaw.map((tenant: any) => ({
            id: tenant.ID,
            name: tenant.Name,
            createdAt: tenant.CreatedAt,
            updatedAt: tenant.UpdatedAt,
          }));
          setTenants(tenantsWithoutStatus);

          console.log("Tenants fetched:", tenantsWithoutStatus);
        }

      } catch (error) {
        console.error("Failed to fetch tenants:", error);
        showToast("Failed to fetch tenants", "error");
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

  const handleAddTenant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTenantName.trim()) return;

    try {
      const token = sessionStorage.getItem("authToken");
      if (!token) {
        throw new Error("Authentication required");
      }

      // Create tenant via API
      const response = await fetch("https://localhost/api/v1/tenants", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify({
          name: newTenantName.trim()
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to create tenant");
      }

      const newTenant = await response.json();
      
      // Add to local state with proper ID from backend
      const tenantData: Tenant = {
        id: newTenant.ID || newTenant.id,
        name: newTenant.Name || newTenant.name,
        createdAt: newTenant.CreatedAt || newTenant.createdAt || new Date().toISOString(),
      };

      setTenants((prev) => [...prev, tenantData]);
      setNewTenantName("");
      setIsAddModalOpen(false);
      
      // Navigate to last page to show new tenant
      const newTotalPages = Math.ceil((filteredTenants.length + 1) / ITEMS_PER_PAGE);
      setCurrentPage(newTotalPages);

      showToast(`Tenant "${tenantData.name}" created successfully`, "success");

    } catch (error) {
      console.error("❌ Failed to create tenant:", error);
      showToast(`Failed to create tenant: ${error instanceof Error ? error.message : "Unknown error"}`, "error");
    }
  };

  const confirmRemoveTenant = (tenant: Tenant) => {
    setTenantToRemove(tenant);
    setVerificationError("");
    setRemoveModalOpen(true);
  };

  const handleRemoveTenant = async () => {
    if (!tenantToRemove) {
      setVerificationError("No tenant selected");
      return;
    }

    setIsRemoving(true);
    setVerificationError("");

    try {
      const token = sessionStorage.getItem("authToken");
      if (!token) {
        throw new Error("Authentication required");
      }

      console.log("🗑️ Starting tenant removal process for:", tenantToRemove.name);

      // Step 1: Fetch all users for this tenant
      console.log("📋 Fetching users for tenant:", tenantToRemove.id);
      
      const usersResponse = await axios.get(`https://localhost/api/v1/users?tenantId=${tenantToRemove.id}`, {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      const usersData = usersResponse.data as { users?: any[] } | any[];
      const users = Array.isArray(usersData)
        ? usersData
        : usersData.users || [];
      console.log(`👥 Found ${users.length} users in tenant "${tenantToRemove.name}"`);

      // Step 2: Delete all users in the tenant
      if (users.length > 0) {
        console.log("🗑️ Removing users from tenant...");
        
        const deletionPromises = users.map(async (user: any) => {
          try {
            const userId = user.ID || user.id || user.userId;
            if (!userId) {
              console.warn("⚠️ User missing ID, skipping:", user);
              return;
            }

            const deleteResponse = await axios.delete(`https://localhost/api/v1/users/${userId}`, {
              headers: {
                Authorization: `Bearer ${token}`,
                'Content-Type': 'application/json'
              }
            });

            console.log(`✅ Deleted user ${user.FullName || user.fullName || user.email || userId}`);
            return deleteResponse;
          } catch (error: any) {
            console.error(`❌ Failed to delete user ${user.FullName || user.email}:`, error);
            throw new Error(`Failed to delete user ${user.FullName || user.email}: ${error.response?.data?.message || error.message}`);
          }
        });

        // Wait for all user deletions to complete
        await Promise.all(deletionPromises);
        console.log("✅ All users deleted successfully");
      }

      console.log("🗑️ Tenant removal process completed");

      // Success - update local state
      setTenants((prev) => prev.filter((t) => t.id !== tenantToRemove.id));
      setRemoveModalOpen(false);
      setTenantToRemove(null);
      
      // Adjust pagination if needed
      if ((filteredTenants.length - 1) <= (currentPage - 1) * ITEMS_PER_PAGE) {
        setCurrentPage(Math.max(1, currentPage - 1));
      }

      // Show success toast instead of alert
      showToast(`Tenant "${tenantToRemove.name}" and all ${users.length} associated users have been successfully removed.`, "success");

    } catch (error) {
      console.error("❌ Failed to remove tenant:", error);
      setVerificationError(error instanceof Error ? error.message : "Failed to remove tenant");
      showToast(`Failed to remove tenant: ${error instanceof Error ? error.message : "Unknown error"}`, "error");
    } finally {
      setIsRemoving(false);
    }
  };

  const btnBase =
    "px-3 py-1.5 rounded-md font-medium text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition";

  return (
    <div className="min-h-screen bg-background text-foreground p-6">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-semibold">Tenants</h1>
        </div>


  {/* Search */}
  <div className="mb-4 flex items-center justify-between gap-4">
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
    <button
      onClick={() => navigate(-1)}
      className="flex items-center gap-2 text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors border border-border hover:bg-muted"
    >
      <ArrowLeft className="w-4 h-4" />
      Back
    </button>  
  </div>
   

        {/* Tenants List */}
        <div className="space-y-4">
          {filteredTenants.length === 0 ? (
            <p className="text-muted-foreground text-center">
              No tenants found.
            </p>
          ) : (
            paginatedTenants.map((tenant) => {
              if (!tenant.name || !tenant.createdAt) return null;

              return (
                <div key={tenant.id} className="flex justify-between items-center border rounded-xl p-4 shadow-md bg-card">
                  <div>
                    <p className="text-lg font-medium">{tenant.name}</p>
                    <p className="text-sm text-muted-foreground">
                      Created: {tenant.createdAt ? new Date(tenant.createdAt).toLocaleString() : "Unknown"}
                    </p>
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={() => confirmRemoveTenant(tenant)}
                      className={`${btnBase} bg-destructive text-destructive-foreground hover:bg-destructive/80`}
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

        {/* Add Tenant Modal */}
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
                className="w-full px-3 py-2 rounded-md border border-input bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
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
                className={`${btnBase} bg-primary text-primary-foreground hover:bg-primary/90`}
              >
                Save
              </button>
            </div>
          </form>
        </Modal>

        {/* Remove Tenant Confirmation Modal */}
        <Modal isOpen={removeModalOpen} onClose={() => setRemoveModalOpen(false)}>
          <h2 className="text-xl font-semibold mb-4 text-destructive">⚠️ Confirm Tenant Removal</h2>
          
          <div className="space-y-4">
            <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-md">
              <p className="text-sm text-destructive font-medium mb-2">
                This action will permanently:
              </p>
              <ul className="text-sm text-destructive space-y-1 ml-4">
                <li>• Remove tenant: <strong>{tenantToRemove?.name}</strong></li>
                <li>• Delete all users associated with this tenant</li>
                <li>• Remove all data and cannot be undone</li>
              </ul>
            </div>

            {verificationError && (
              <p className="text-sm text-destructive">{verificationError}</p>
            )}
          </div>

          <div className="flex justify-end gap-3 mt-6">
            <button
              onClick={() => {
                setRemoveModalOpen(false);
                setVerificationError("");
              }}
              className={`${btnBase} border border-border bg-muted hover:bg-muted/70 text-foreground`}
              disabled={isRemoving}
            >
              Cancel
            </button>
            <button
              onClick={handleRemoveTenant}
              disabled={isRemoving}
              className={`${btnBase} bg-destructive text-destructive-foreground hover:bg-destructive/80 disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              {isRemoving ? "Removing..." : "Confirm Removal"}
            </button>
          </div>
        </Modal>

        {/* Toast Notification */}
        {toast && (
          <Toast 
            message={toast.message} 
            type={toast.type} 
            onClose={() => setToast(null)} 
          />
        )}
      </div>
    </div>
  );
};