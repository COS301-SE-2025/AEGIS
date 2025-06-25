import React, { useState } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { Button } from '../../components/ui/button';
import { Input } from '../../components/ui/input';
import { MailIcon, Share2Icon } from 'lucide-react';

export function ShareCaseForm(): JSX.Element {
  const { caseId } = useParams<{ caseId: string }>();
  const location = useLocation();
  const navigate = useNavigate();

  const caseName = location.state?.caseName || 'Untitled Case';

  const [form, setForm] = useState({
    senderName: "",
    recipientEmail: "",
  });

  const handleChange = (field: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [field]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const payload = {
      ...form,
      caseId: caseId,
    };

    try {
      console.log(`Sharing case ${caseId} with ${form.recipientEmail}...`);
      console.log("Payload:", payload);

      // TODO: Integrate API call here

      alert(`Case "${caseName}" shared successfully with ${form.recipientEmail}!`);
      navigate('/dashboard');
    } catch (error) {
      console.error("Error sharing case:", error);
      alert("Failed to share the case. Please try again.");
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-xl w-full bg-card border border-border p-6 rounded-2xl shadow-xl font-mono">
        <h1 className="text-3xl font-bold text-primary mb-6 flex items-center gap-2">
          <Share2Icon size={28} /> Share Case
        </h1>

        {caseName && (
          <p className="mb-4 text-sm text-muted-foreground">
            Sharing: <span className="text-primary font-bold">{caseName}</span>
          </p>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block mb-1 text-sm text-foreground">Your Name</label>
            <Input
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="e.g. Adam Forensics"
              value={form.senderName}
              onChange={handleChange("senderName")}
              required
            />
          </div>

          <div>
            <label className="block mb-1 text-sm text-foreground">Recipient Email</label>
            <Input
              type="email"
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="e.g. analyst@partner.com"
              value={form.recipientEmail}
              onChange={handleChange("recipientEmail")}
              required
            />
          </div>

          <div className="flex flex-wrap gap-4 pt-4">
            <Button
              type="button"
              variant="outline"
              className="border-destructive text-destructive hover:bg-destructive/10"
              onClick={() => navigate(-1)}
            >
              Cancel
            </Button>

            <Button
              type="submit"
              className="bg-cyan-400 text-cyan-600 hover:bg-primary/90"
            >
              <MailIcon size={18} className="mr-2" />
              Send Invite
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
