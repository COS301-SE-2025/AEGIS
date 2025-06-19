//Form to share  a case with an external user/collaborator
import React, { useState } from 'react';
import { useParams, useLocation, useNavigation, useNavigate } from 'react-router-dom';
import { Button } from '../../components/ui/button';
import { Input } from '../../components/ui/input';
import { MailIcon, Share2Icon } from 'lucide-react';

export function ShareCaseForm(): JSX.Element {
    const { caseId } = useParams<{ caseId: string }>();
    const location = useLocation();
    const navigate = useNavigate();

    const caseName = location.state?.caseName || 'Untitled Case'; //retrieve case name from location state or default to 'Untitled Case'

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
            caseId: caseId
        };

        // Simulate API call to share the case
        try {
            console.log(`Sharing case ${caseId} with ${form.recipientEmail}...`);
            console.log("Payload:", payload);

            //TODO: Send payload to backend

            alert(`Case "${caseName}" shared successfully with ${form.recipientEmail}!`);
            //successfully shared the case, redirect to dashboard
            navigate('/dashboard');

        } catch (error) {

            console.error("Error sharing case:", error);
            alert("Failed to share the case. Please try again.");
        }
    };

     return (
    <div className="min-h-screen bg-zinc-900 text-white flex items-center justify-center p-6">
      <div className="max-w-xl w-full bg-zinc-900 border border-zinc-700 p-6 rounded-2xl shadow-xl font-mono">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
          <Share2Icon size={28} /> Share Case
        </h1>

        {caseName && (
          <p className="mb-4 text-sm text-zinc-400">
            Sharing: <span className="text-cyan-300 font-bold">{caseName}</span>
          </p>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block mb-1 text-sm">Your Name</label>
            <Input
              className="bg-zinc-800 border-zinc-600 text-white"
              placeholder="e.g. Adam Forensics"
              value={form.senderName}
              onChange={handleChange("senderName")}
              required
            />
          </div>

          <div>
            <label className="block mb-1 text-sm">Recipient Email</label>
            <Input
              type="email"
              className="bg-zinc-800 border-zinc-600 text-white"
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
              className="bg-zinc-800 border-red-500 text-red-400 hover:bg-red-800"
              onClick={() => navigate(-1)}
            >
              Cancel
            </Button>

            <Button
              type="submit"
              className="bg-cyan-600 hover:bg-cyan-700 text-white"
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

