import { useState } from "react";
import { useNavigate } from "react-router-dom";

export const TermsAndConditionsPage = () => {
  const [accepted, setAccepted] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = () => {
    fetch("/api/auth/accept-terms", { method: "POST" })
      .then(() => navigate("/login"))
      .catch(() => alert("Something went wrong."));
  };

  return (
    <div className="min-h-screen bg-background text-foreground p-8 font-mono">
      <div className="max-w-3xl mx-auto bg-muted p-6 rounded-xl shadow-lg border border-primary">
        <h1 className="text-2xl text-green-400 mb-4 font-bold">AEGIS Terms & Conditions</h1>

        <div className="h-64 overflow-y-scroll p-4 bg-background border border-border rounded">
          <p className="text-muted-foreground text-sm leading-relaxed">
            By using AEGIS, you agree to abide by our cybersecurity standards, policies and practices.
            Your personal data is handled in accordance with POPIA/GDPR and will not be shared
            without your consent. Activity may be logged for integrity verification.
            <br /><br />
            You have the right to access, update, or delete your information. Accepting these
            terms is required to use the platform. Your consent can be revoked at any time if you so wish.
          </p>
        </div>

        <div className="flex items-center mt-6">
          <input
            type="checkbox"
            id="accept"
            checked={accepted}
            onChange={() => setAccepted(!accepted)}
            className="mr-2 accent-primary"
          />
          <label htmlFor="accept" className="text-sm">
            I have read and accept the Terms & Conditions
          </label>
        </div>

        <div className="mt-4 flex gap-4">
          <button
            disabled={!accepted}
            onClick={handleSubmit}
            className={`px-6 py-2 rounded-xl font-semibold transition ${
              accepted
                ? "bg-green-500 hover:bg-green-600 text-black"
                : "bg-zinc-600 text-zinc-400 cursor-not-allowed"
            }`}
          >
            Accept & Continue
          </button>
          <button
            onClick={() => alert("You must accept the terms to proceed.")}
            className="px-6 py-2 bg-red-500 hover:bg-red-600 text-black rounded-xl"
          >
            Reject
          </button>
        </div>
      </div>
    </div>
  );
};
