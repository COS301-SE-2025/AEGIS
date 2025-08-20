import {
  Bell,
  Settings,
}from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";

export const TenantAdminDashboard = () => {
  const navigate = useNavigate();
  const storedUser = sessionStorage.getItem("user");
  const user = storedUser ? JSON.parse(storedUser) : null;
    const [, setProfile] = useState<{ name: string; email: string; role: string; image: string } | null>(null);

  const displayName = user?.name || user?.email?.split("@")[0] || "Agent User";
  const initials = displayName
    .split(" ")
    .map((part: string) => part[0])
    .join("")
    .toUpperCase();
    useEffect(() => {
        const fetchProfile = async () => {
          try {
            const token = sessionStorage.getItem("authToken");
            const res = await fetch(`http://localhost:8080/api/v1/profile/${user?.id}`, {
              headers: {
                Authorization: `Bearer ${token}`,
              },
            });
    
            if (!res.ok) throw new Error("Failed to load profile");
    
            const result = await res.json();
    
            // Update both the state and sessionStorage
            setProfile({
              name: result.data.name,
              email: result.data.email,
              role: result.data.role,
              image: result.data.image_url,
            });
    
            // Update sessionStorage
            sessionStorage.setItem(
              "user",
              JSON.stringify({
                ...user,
                name: result.data.name,
                email: result.data.email,
                image_url: result.data.image_url,
              })
            );
          } catch (err) {
            console.error("Error fetching profile:", err);
          }
        };
    
        if (user?.id) fetchProfile();
      }, [user?.id]);
    

  return (
    <div className="min-h-screen px-8 py-10 bg-background text-foreground transition-colors">
      <div className="max-w-6xl mx-auto space-y-6">
      {/* Flex container for heading and navbar */}
    <div className="flex items-center justify-between border-b border-muted pb-4">
      {/* Left: Heading */}
      <h1 className="text-3xl font-bold flex items-center gap-2 text-foreground">
        Dashboard
      </h1>

      {/* Right: Navigation buttons */}
      <div className="flex gap-4">
          <Link to="/settings">
          <button className="p-2 text-muted-foreground hover:text-white transition-colors">
          <Settings className="w-6 h-6" />
          </button>
          </Link>
          <Link to="/notifications">
          <button className="p-2 text-muted-foreground hover:text-white transition-colors">
          <Bell className="w-6 h-6" />
          </button>
          </Link>
          <Link to="/profile">
          {user?.image_url ? (
            <img
              src={
              user.image_url.startsWith("http") || user.image_url.startsWith("data:")
              ? user.image_url
              : `http://localhost:8080${user.image_url}`
              }
            alt="Profile"
            className="w-10 h-10 rounded-full object-cover"
            />
            ) : (
           <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
             <span className="text-foreground font-medium text-sm">{initials}</span>
            </div>
          )}
          </Link>


      </div>
    </div>

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
