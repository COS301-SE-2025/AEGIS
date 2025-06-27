import React from "react";

export const About = () => {
  return (
    <div className="min-h-screen bg-gray-900 text-white px-6 py-20">
      <div className="max-w-5xl mx-auto space-y-16">
        {/* Heading */}
        <section className="text-center">
          <h1 className="text-4xl font-bold mb-4 text-blue-400">
            About AEGIS
          </h1>
          <p className="text-gray-300 text-lg max-w-3xl mx-auto">
            AEGIS (Automated Evidence Generation and Integrity System) is an advanced cybersecurity and digital forensics platform designed to support secure, efficient, and collaborative investigation of cyber incidents.
          </p>
        </section>

        {/* Purpose */}
        <section>
          <h2 className="text-2xl font-semibold text-white mb-3">Why AEGIS?</h2>
          <p className="text-gray-300">
            Digital Forensics and Incident Response (DFIR) teams face increasing challenges: maintaining the integrity of digital evidence, enabling real-time collaboration, and producing clear and auditable case histories. AEGIS addresses these needs by combining automation, security, and collaboration tools into a single unified platform.
          </p>
        </section>

        {/* Core Capabilities */}
        <section>
          <h2 className="text-2xl font-semibold text-white mb-6">Core Capabilities</h2>
          <div className="grid md:grid-cols-2 gap-8 text-gray-300">
            <div>
              <h3 className="font-semibold text-blue-400 mb-2">🧾 Case Management</h3>
              <ul className="list-disc list-inside space-y-1">
                <li>Create and track cases with role-based access</li>
                <li>Visual timeline for tracking incident events</li>
                <li>Automatic logging of all evidence actions</li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-blue-400 mb-2">🧪 Collaborative Evidence Analysis</h3>
              <ul className="list-disc list-inside space-y-1">
                <li>Support for logs, images, packet captures, and more</li>
                <li>Real-time multi-user collaboration and chat</li>
                <li>Threaded annotations on specific evidence</li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-blue-400 mb-2">🔐 Secure Communication</h3>
              <ul className="list-disc list-inside space-y-1">
                <li>End-to-end encryption for messages and files</li>
                <li>Encrypted storage for all case data</li>
                <li>Secure file sharing among team members</li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-blue-400 mb-2">🛡️ Chain of Custody</h3>
              <ul className="list-disc list-inside space-y-1">
                <li>Document collection methods and timestamps</li>
                <li>Track all forensic steps and tool usage</li>
                <li>Generate detailed investigation reports</li>
              </ul>
            </div>
          </div>
        </section>

        {/* Use Cases */}
        <section>
          <h2 className="text-2xl font-semibold text-white mb-3">Who Is AEGIS For?</h2>
          <p className="text-gray-300">
            AEGIS is ideal for cybersecurity professionals, DFIR teams, enterprise security departments, and digital investigators who need a secure and efficient platform for managing cyber incidents and forensic investigations.
          </p>
        </section>

        {/* Contact */}
        <section className="text-center pt-10 border-t border-gray-700">
          <h3 className="text-xl font-semibold text-white mb-2">Contact Us</h3>
          <p className="text-gray-400">
            📧 capstone.incidentintel@gmail.com
          </p>
        </section>
      </div>
    </div>
  );
};
