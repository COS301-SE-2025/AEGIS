import { useState, useEffect } from "react";
import { Input } from "../../components/ui/input";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "../../components/ui/select";
import { Pagination } from "../../components/ui/pagination";

interface Team {
  id: string;
  name: string;
  manager: string;
  members: number;
  status: "Active" | "Inactive";
}

const mockTeams: Team[] = Array.from({ length: 23 }).map((_, i) => ({
  id: `team-${i + 1}`,
  name: `DFIR Team ${i + 1}`,
  manager: `Manager ${i + 1}`,
  members: Math.floor(Math.random() * 10) + 3,
  status: Math.random() > 0.5 ? "Active" : "Inactive",
}));

export const TeamsPage = () => {
  const [teams, setTeams] = useState<Team[]>(mockTeams);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<"All" | "Active" | "Inactive">("All");
  const [page, setPage] = useState(1);
  const itemsPerPage = 6;

  useEffect(() => {
    let filtered = mockTeams;

    if (search) {
      filtered = filtered.filter((t) =>
        t.name.toLowerCase().includes(search.toLowerCase())
      );
    }

    if (statusFilter !== "All") {
      filtered = filtered.filter((t) => t.status === statusFilter);
    }

    setTeams(filtered);
    setPage(1);
  }, [search, statusFilter]);

  const totalPages = Math.ceil(teams.length / itemsPerPage);
  const paginated = teams.slice((page - 1) * itemsPerPage, page * itemsPerPage);

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
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as any)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="All">All Statuses</SelectItem>
              <SelectItem value="Active">Active</SelectItem>
              <SelectItem value="Inactive">Inactive</SelectItem>
            </SelectContent>
          </Select>
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
                <div
                  className={`text-sm font-medium ${
                    team.status === "Active" ? "text-green-500" : "text-red-500"
                  }`}
                >
                  {team.status}
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
