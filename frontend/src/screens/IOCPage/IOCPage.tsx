import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import axios from "axios";
import { Button } from "../../components/ui/button";
import { toast } from "react-hot-toast";

interface Case {
  id: string;
  name: string;
}

interface IOC {
  id: string; // changed to string
  type: string;
  value: string;
  created_at: string;
}

const iocTypes = ["IP", "Email", "Domain", "Hash", "URL"];

export const IOCPage = () => {
  const { case_id } = useParams<{ case_id: string }>(); // match route param naming
  const navigate = useNavigate();

  const [caseName, setCaseName] = useState("");
  const [iocs, setIocs] = useState<IOC[]>([]);
  const [type, setType] = useState("IP");
  const [value, setValue] = useState("");
  const [loading, setLoading] = useState(false);

  const token = sessionStorage.getItem("authToken");

  useEffect(() => {
    async function fetchData() {
      if (!case_id) return;

      if (!token) {
        toast.error("You are not authenticated. Please log in.");
        return;
      }

      try {
        setLoading(true);

        // GET Case details
        const caseRes = await axios.get<Case>(`http://localhost:8080/api/v1/cases/${case_id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        setCaseName(caseRes.data.name);

        // GET IOCs for case
        const iocRes = await axios.get<IOC[]>(`http://localhost:8080/api/v1/cases/${case_id}/iocs`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (Array.isArray(iocRes.data)) {
          setIocs(iocRes.data);
        } else {
          setIocs([]);
        }
      } catch (err) {
        toast.error("Failed to load case or IOCs");
        setIocs([]);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [case_id, token]);

  async function handleAddIOC() {
    if (!value.trim()) {
      toast.error("Please enter a value for the IOC.");
      return;
    }

    if (!case_id) {
      toast.error("Invalid case ID.");
      return;
    }
    if (!token) {
      toast.error("You are not authenticated. Please log in.");
      return;
    }

    try {
      setLoading(true);

      const res = await axios.post<IOC>(
        `http://localhost:8080/api/v1/cases/${case_id}/iocs`,
        { type, value },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      setIocs((prev) => [...prev, res.data]);
      setValue("");
      toast.success("IOC added successfully!");
    } catch {
      toast.error("Failed to add IOC.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="flex items-center gap-4 mb-8">
        <button
          onClick={() => navigate(-1)}
          className="text-blue-600 hover:text-blue-800 focus:outline-none"
          aria-label="Go back"
        >
          ← Back
        </button>
        <h1 className="text-3xl font-bold text-background">IOCs for Case: {caseName}</h1>
      </div>

      <div className="mb-8 border border-gray-300 rounded-lg p-6 bg-popover">
        <h2 className="text-xl font-semibold mb-4">Add New IOC</h2>
        <div className="flex flex-col sm:flex-row sm:items-center gap-4">
          <select
            value={type}
            onChange={(e) => setType(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-2 text-background focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            {iocTypes.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
          <input
            type="text"
            placeholder="Enter IOC value"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-background placeholder-muted-background focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <Button onClick={handleAddIOC} disabled={loading}>
            {loading ? "Adding..." : "Add IOC"}
          </Button>
        </div>
      </div>

      <div>
        <h2 className="text-xl font-semibold mb-4">Existing IOCs</h2>

        {loading && iocs.length === 0 ? (
          <p>Loading IOCs...</p>
        ) : !Array.isArray(iocs) || iocs.length === 0 ? (
          <p className="text-muted-background">No IOCs found for this case.</p>
        ) : (
          <ul className="space-y-3">
            {iocs.map(({ id, type, value, created_at }) => (
              <li
                key={id}
                className="flex justify-between items-center border border-gray-300 rounded-md px-4 py-2 bg-popover"
              >
                <div>
                  <span className="font-semibold mr-2">{type}:</span>
                  <span>{value}</span>
                </div>
                <time
                  className="text-xs text-muted-background"
                  dateTime={created_at}
                  title={new Date(created_at).toLocaleString()}
                >
                  {new Date(created_at).toLocaleDateString()}
                </time>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};
