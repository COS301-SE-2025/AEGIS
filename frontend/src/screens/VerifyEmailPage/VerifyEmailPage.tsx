import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { LoaderPinwheelIcon } from "lucide-react";

export const VerifyEmailPage = () => {
    const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();

    useEffect(() => {
        const token = searchParams.get("token");
        if (!token) {
            setStatus("error");
            return;
        }

        fetch("/api/auth/verify-email", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ token }),
        })
            .then((response) => {
                if (response.ok) {
                    setStatus("success");
                } else {
                    setStatus("error");
                }
            })
            .catch(() => {
                setStatus("error");
            });
    }, [searchParams]);

    useEffect(() => {
        if (status === "success") {
            const timer = setTimeout(() => {
                navigate("/login");
            }, 2000);
            return () => clearTimeout(timer);
        }
    }, [status, navigate]);

    return (
    <div className="min-h-screen bg-black text-green-400 flex flex-col items-center justify-center font-mono p-4">
      {status === "loading" && (
        <>
          <LoaderPinwheelIcon />
          <p className="mt-4">Verifying your email...</p>
        </>
      )}
      {status === "success" && (
        <p className="text-lg animate-pulse">✅ Email verified. Redirecting to T&Cs...</p>
      )}
      {status === "error" && (
        <p className="text-red-500 text-lg">❌ Invalid or expired token.</p>
      )}
    </div>
  );
}