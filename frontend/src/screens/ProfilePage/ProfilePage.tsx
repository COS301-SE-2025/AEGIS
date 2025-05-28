import { useState, useRef } from "react";
import { Pencil, Save, Upload, User, Mail, Shield } from "lucide-react";

export  const ProfilePage= () => {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [isEditing, setIsEditing] = useState(false);
  const [profile, setProfile] = useState({
    name: "Agent Carter",
    email: "carter@aegis.com",
    role: "Incident Responder",
    image: null as string | null,
  });

  const toggleEdit = () => setIsEditing(!isEditing);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setProfile({ ...profile, [e.target.name]: e.target.value });
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const reader = new FileReader();
      reader.onload = () => {
        setProfile({ ...profile, image: reader.result as string });
      };
      reader.readAsDataURL(e.target.files[0]);
    }
  };

  return (
    <div className="bg-black min-h-screen text-white p-10">
      <h1 className="text-3xl font-bold mb-8">Profile</h1>

      <div className="bg-gray-900 p-6 rounded-lg max-w-xl mx-auto shadow-lg space-y-6">
        {/* Profile Picture */}
        <div className="flex flex-col items-center space-y-3">
          <div className="relative w-24 h-24">
            <img
              src={
                profile.image ||
                "https://ui-avatars.com/api/?name=Agent+Carter&background=0D8ABC&color=fff"
              }
              alt="Profile"
              className="w-24 h-24 rounded-full object-cover border-4 border-gray-800"
            />
            {isEditing && (
              <button
                onClick={() => fileInputRef.current?.click()}
                className="absolute bottom-0 right-0 bg-blue-600 hover:bg-blue-500 p-1 rounded-full"
              >
                <Upload className="w-4 h-4 text-white" />
              </button>
            )}
            <input
              type="file"
              accept="image/*"
              ref={fileInputRef}
              className="hidden"
              onChange={handleImageChange}
            />
          </div>
        </div>

        {/* Name */}
        <div className="flex items-center gap-4">
          <User className="w-6 h-6 text-gray-400" />
          {isEditing ? (
            <input
              type="text"
              name="name"
              value={profile.name}
              onChange={handleChange}
              className="bg-gray-800 p-2 rounded text-white w-full"
            />
          ) : (
            <p className="text-lg">{profile.name}</p>
          )}
        </div>

        {/* Email */}
        <div className="flex items-center gap-4">
          <Mail className="w-6 h-6 text-gray-400" />
          {isEditing ? (
            <input
              type="email"
              name="email"
              value={profile.email}
              onChange={handleChange}
              className="bg-gray-800 p-2 rounded text-white w-full"
            />
          ) : (
            <p className="text-lg">{profile.email}</p>
          )}
        </div>

            {/* Role */}
    <div className="flex items-center gap-4">
    <Shield className="w-6 h-6 text-gray-400" />
    {isEditing ? (
        <input
        type="text"
        name="role"
        value={profile.role}
        disabled //  makes the input read-only
        className="bg-gray-800 p-2 rounded text-white w-full opacity-70 cursor-not-allowed"
        />
    ) : (
        <p className="text-lg">{profile.role}</p>
    )}
    </div>

        {/* Role Description */}

        {/* Actions */}
        <div className="flex justify-end mt-4">
          <button
            onClick={toggleEdit}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 px-4 py-2 rounded-lg transition"
          >
            {isEditing ? <Save className="w-4 h-4" /> : <Pencil className="w-4 h-4" />}
            {isEditing ? "Save Changes" : "Edit Profile"}
          </button>
        </div>
      </div>
    </div>
  );
}
