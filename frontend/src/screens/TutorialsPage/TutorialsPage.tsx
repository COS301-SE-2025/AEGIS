import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router-dom";

const tutorials = [
  {
    title: "Getting Started with AEGIS",
    description: "Log in to your AEGIS account.",
    videoUrl: "https://www.youtube.com/embed/n5-JYfQzaoI",
  },
  {
    title: "Creating a New Case",
    videoUrl: "https://www.youtube.com/embed/qoT7659bk-M",
    description: "Learn how to create a new case in AEGIS.",
  },
  {
    title: "Viewing Case Details",
    videoUrl: "https://www.youtube.com/embed/tgwSBYAoj5w",
    description: "Learn how to view case details in AEGIS.",
  },
  {
    title: "Creating a New Group Chat",
    videoUrl: "https://www.youtube.com/embed/qV8x_ISiD68",
    description: "Learn how to create a new group chat in AEGIS.",
  },
];

export const TutorialsPage = () => {
  const navigate = useNavigate();

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
        <div className="max-w-6xl mx-auto">
          <h1 className="text-4xl font-bold mb-10 text-center">AEGIS Tutorials</h1>

          <div className="grid md:grid-cols-2 gap-10">
            {tutorials.map((tutorial, index) => (
              <div
                key={index}
                className="bg-gray-800 rounded-xl shadow-lg overflow-hidden border border-gray-700"
              >
                <div className="aspect-video">
                  <iframe
                    className="w-full h-full"
                    src={tutorial.videoUrl}
                    title={tutorial.title}
                    frameBorder="0"
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                    allowFullScreen
                  ></iframe>
                </div>
                <div className="p-6">
                  <h2 className="text-xl font-semibold mb-2">{tutorial.title}</h2>
                  <p className="text-gray-300 text-sm">{tutorial.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
