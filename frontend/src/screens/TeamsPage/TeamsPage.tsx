import { useEffect, useState } from "react";
import { Input } from "../../components/ui/input";
import { Pagination } from "../../components/ui/pagination";
import { jwtDecode } from "jwt-decode";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router-dom";

interface Team {
  id: string;
  name: string;
  manager: string;
  members: number;
  status: "Active" | "Inactive";
}

export const TeamsPage = () => {
  const [teams, setTeams] = useState<Team[]>([]);
  const [filteredTeams, setFilteredTeams] = useState<Team[]>([]);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const itemsPerPage = 6;
  const statuses = ["Active", "Inactive"];
  const navigate = useNavigate();

  useEffect(() => {
    const fetchTeams = async () => {
      const token = sessionStorage.getItem("authToken");
      if (!token) {
        console.error("No token found");
        return;
      }

      let tenantId: string | null = null;
      try {
        const decoded: any = jwtDecode(token);
        tenantId = decoded.tenant_id;
      } catch (e) {
        console.error("Failed to decode token", e);
        return;
      }

      if (!tenantId) {
        console.error("No tenant ID found in token");
        return;
      }

      try {
        const res = await fetch(
          `https://localhost/api/v1/teams?tenant_id=${tenantId}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        if (!res.ok) {
          const errPayload = await res.json();
          console.error("Failed to fetch teams:", errPayload.message);
          return;
        }
        const teamsData = await res.json();
        const mapped = teamsData.map((team: any) => ({
          id: team.id,
          name: team.name,
          manager: team.manager, // Or dynamically derive from backend later
          members: 0, // Placeholder for now
          status:
            statuses[Math.floor(Math.random() * statuses.length)], // Random status for demo
        }));
        setTeams(mapped);
        setFilteredTeams(teamsData);
      } catch (err) {
        console.error("Error fetching teams:", err);
      }
    };

    fetchTeams();
  }, []);

  useEffect(() => {
    const filtered = teams.filter(
      (t) =>
        typeof t.name === "string" &&
        t.name.toLowerCase().includes(search.toLowerCase())
    );
    setFilteredTeams(filtered);
    setPage(1);
  }, [search, teams]);

  const totalPages = Math.ceil(filteredTeams.length / itemsPerPage);
  const paginated = filteredTeams.slice(
    (page - 1) * itemsPerPage,
    page * itemsPerPage
  );

  // Handle back navigation
  const handleBack = () => {
    navigate(-1); // Go back to previous page
  };

  return (
    <div className="min-h-screen bg-background text-foreground p-6">
      <div className="max-w-6xl mx-auto space-y-6">
        <h1 className="text-3xl font-bold">DFIR Teams</h1>

        <div className="flex flex-col md:flex-row items-center justify-between gap-4">
          <Input
            placeholder="Search by team name..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-sm"
          />

          <button
            onClick={handleBack}
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors border border-border hover:bg-muted"
          >
            <ArrowLeft className="w-4 h-4" />
            Back
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 pt-2">
          {paginated.length === 0 ? (
            <div className="col-span-full text-center text-muted-foreground py-10">
              No teams found.
            </div>
          ) : (
            paginated.map((team) => (
              <div
                key={team.id}
                className="bg-card rounded-2xl shadow p-4 border border-border space-y-2"
              >
                <div className="text-lg font-semibold">{team.name}</div>
                <div className="text-sm text-muted-foreground">
                  Manager: {team.manager}
                </div>
                <div className="text-sm text-muted-foreground">
                  Members: {team.members}
                </div>
                <div>
                  <span
                    className={`inline-block px-2 py-1 rounded-full text-xs font-medium ${
                      team.status === "Active"
                        ? "bg-success text-success-foreground"
                        : team.status === "Inactive"
                        ? "bg-destructive text-destructive-foreground"
                        : "bg-warning text-warning-foreground"
                    }`}
                  >
                    {team.status}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>

        <div className="flex justify-center pt-6">
          <Pagination page={page} totalPages={totalPages} onChange={setPage} />
        </div>
      </div>
    </div>
  );
};
