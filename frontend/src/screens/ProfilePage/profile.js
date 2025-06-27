const updateProfile = async ({ name, email, imageFile }) => {
  const formData = new FormData();
  formData.append("name", name);
  formData.append("email", email);
  if (imageFile) {
    formData.append("profile_picture", imageFile);
  }

  const response = await fetch("http://localhost:8080/api/v1/profile/update", {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    throw new Error("Failed to update profile");
  }

  return await response.json();
};
export default updateProfile;

