import { useNavigate } from "react-router-dom";

function ShareButton({ caseId, caseName }: { caseId: string; caseName: string }) {
  const navigate = useNavigate();

  return (
    <button
      onClick={() => navigate(`/cases/${caseId}/share`, {
        state: { caseName },
      })}
    >
      Share This Case
    </button>
  );
}
export { ShareButton};