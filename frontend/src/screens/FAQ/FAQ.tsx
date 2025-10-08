import { useState } from "react";
import { ChevronDown, ChevronUp, ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router-dom";

interface FAQItem {
  question: string;
  answer: string;
}

const faqs: FAQItem[] = [
  {
    question: "What is AEGIS?",
    answer:
      "AEGIS (Automated Evidence Generation and Integrity System) is a platform for digital forensics and incident response teams to securely manage cases, collaborate in real-time, and preserve the integrity of digital evidence.",
  },
  {
    question: "How does AEGIS protect the integrity of evidence?",
    answer:
      "AEGIS enforces a strict chain-of-custody process by automatically logging all actions performed on evidence. This ensures a tamper-proof audit trail from acquisition to reporting.",
  },
  {
    question: "Can multiple users work on a case simultaneously?",
    answer:
      "Yes. AEGIS supports real-time collaboration, including commenting, threaded discussions, and annotations on evidence to enable distributed teams to work efficiently together.",
  },
  {
    question: "What types of evidence formats does AEGIS support?",
    answer:
      "AEGIS supports various formats such as log files, images, packet captures, and disk images. These are presented using visual tools to aid interpretation and analysis.",
  },
  {
    question: "How does AEGIS ensure secure communication?",
    answer:
      "All communication within AEGIS is end-to-end encrypted. Files and messages are encrypted both in transit and at rest to prevent unauthorized access.",
  },
  {
    question: "Can external parties be granted access to a case?",
    answer:
      "Yes. Administrators can generate time-limited, secure access tokens to share specific cases or files with external collaborators.",
  },
  {
    question: "Does AEGIS support role-based access control?",
    answer:
      "Absolutely. Access to cases, evidence, and investigation stages is controlled through configurable roles assigned to users by administrators.",
  },
  {
    question: "Is it possible to generate investigation reports?",
    answer:
      "Yes. For each case, AEGIS can generate a structured report that documents all evidence handling, investigation steps, and analytical findings.",
  },
];

export const FAQ = () => {
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  const navigate = useNavigate();

  const toggle = (index: number) => {
    setOpenIndex(prev => (prev === index ? null : index));
  };

  const handleBack = () => {
    navigate(-1); // Go back to previous page
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      {/* Back Button - Top Left */}
      <div className="absolute top-6 left-6 z-10">
        <button
          onClick={handleBack}
          className="flex items-center gap-2 text-blue-400 hover:text-white border border-blue-400 hover:border-white px-3 py-2 rounded-lg transition-colors"
          aria-label="Go back"
        >
          <ArrowLeft className="w-5 h-5" />
          <span>Back</span>
        </button>
      </div>

      <div className="px-6 py-20">
        <div className="max-w-5xl mx-auto">
          <h1 className="text-4xl font-bold mb-10 text-center text-blue-400">
            Frequently Asked Questions
          </h1>

          <div className="space-y-6">
            {faqs.map((faq, index) => (
              <div key={index} className="border border-gray-700 rounded-lg overflow-hidden">
                <button
                  onClick={() => toggle(index)}
                  className="w-full px-6 py-4 flex justify-between items-center text-left hover:bg-gray-800 transition"
                  aria-expanded={openIndex === index}
                >
                  <span className="text-lg font-medium">{faq.question}</span>
                  {openIndex === index ? (
                    <ChevronUp className="w-5 h-5 text-blue-400" />
                  ) : (
                    <ChevronDown className="w-5 h-5 text-gray-400" />
                  )}
                </button>
                {openIndex === index && (
                  <div className="px-6 pb-4 text-gray-300">{faq.answer}</div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
