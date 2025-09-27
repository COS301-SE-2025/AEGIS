const updateProfile = async ({ id, name, email, imageBase64 }) => {
  const token = sessionStorage.getItem("authToken");

  const body = {
    id,
    name,
    email,
    imageBase64: imageBase64 || "", // Ensure it's a string, not undefined
  };

  const response = await fetch("https://localhost/api/v1/profile", {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });

  const json = await response.json();

  if (!response.ok) {
    console.error("Update failed:", json);
    throw new Error(json.message || "Profile update failed");
  }

  //console.log("✅ Profile Data:", json);
  return json.data; // Return the updated profile data
};

export default updateProfile;
