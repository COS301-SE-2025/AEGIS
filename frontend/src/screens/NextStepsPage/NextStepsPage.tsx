// src/pages/NextStepsPage.tsx
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "../../components/ui/button";

export default function NextStepsPage() {
  //const { id: caseId } = useParams();
  const caseId = useParams().id || JSON.parse(localStorage.getItem("currentCase") || "{}").id;
  console.log("Case ID from params of next-page:", caseId);

  const navigate = useNavigate();

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6 text-center bg-background text-foreground">
      <div className="bg-card p-8 rounded-xl shadow-xl w-full max-w-md border">
  <h1 className="text-2xl font-bold text-success mb-2">
          ✅ Case #{caseId} Created
        </h1>
        <p className="text-muted-foreground mb-6">
          What would you like to do next?
        </p>

        <div className="space-y-3">
          <Button className="w-full" onClick={() => navigate(`/upload-evidence/${caseId}`)}>
            Upload Evidence
          </Button>
          <Button className="w-full" onClick={() => navigate(`/assign-case-members/${caseId}`)}>
            Assign Members
          </Button>
          <Button
            variant="outline"
            className="w-full"
            onClick={() => navigate("/dashboard")}
          >
            Go to Dashboard
          </Button>
        </div>
      </div>
    </div>
  );
}
