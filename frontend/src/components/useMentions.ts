import { useState, useEffect, useCallback } from 'react';

export interface IOC {
  id: string;
  tenant_id: string;
  case_id: string;
  type: string; // e.g., IP, Email, Domain
  value: string;
  created_at: string;
}

export interface Evidence {
  id: string;
  case_id: string;
  uploaded_by: string;
  tenant_id: string;
  team_id: string;
  filename: string;
  file_type: string;
  ipfs_cid: string;
  file_size: number;
  checksum: string;
  metadata: string;
  uploaded_at: string;
}

export type MentionItem = {
  id: string;
  type: 'ioc' | 'evidence';
  display: string;
  subtitle: string;
  data: IOC | Evidence;
};

interface UseMentionsProps {
  caseId: string;
  apiBaseUrl?: string;
}

export const useMentions = ({ caseId, apiBaseUrl = 'https://localhost/api/v1' }: UseMentionsProps) => {
  const [iocs, setIOCs] = useState<IOC[]>([]);
  const [evidence, setEvidence] = useState<Evidence[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch IOCs for the case
  const fetchIOCs = useCallback(async () => {
    if (!caseId) return;
    
    try {
      setLoading(true);
        const token = sessionStorage.getItem("authToken") || "";
      const response = await fetch(`${apiBaseUrl}/cases/${caseId}/iocs`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch IOCs: ${response.statusText}`);
      }

      const data = await response.json();
      setIOCs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch IOCs');
      console.error('Error fetching IOCs:', err);
    } finally {
      setLoading(false);
    }
  }, [caseId, apiBaseUrl]);

  // Fetch evidence for the case
  const fetchEvidence = useCallback(async () => {
    if (!caseId) return;
    
    try {
      setLoading(true);
      const token = sessionStorage.getItem("authToken") || "";
      const response = await fetch(`${apiBaseUrl}/evidence-metadata/case/${caseId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch evidence: ${response.statusText}`);
      }

      const data = await response.json();
      setEvidence(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch evidence');
      console.error('Error fetching evidence:', err);
    } finally {
      setLoading(false);
    }
  }, [caseId, apiBaseUrl]);

  // Initialize data
  useEffect(() => {
    if (caseId) {
      fetchIOCs();
      fetchEvidence();
    }
  }, [caseId, fetchIOCs, fetchEvidence]);

  // Convert data to mention items
  const getMentionItems = useCallback((query: string = ''): MentionItem[] => {
    const items: MentionItem[] = [];
    
    // Add IOCs
    iocs
      .filter(ioc => 
        !query || 
        ioc.value.toLowerCase().includes(query.toLowerCase()) ||
        ioc.type.toLowerCase().includes(query.toLowerCase())
      )
      .forEach(ioc => {
        items.push({
          id: `ioc-${ioc.id}`,
          type: 'ioc',
          display: `${ioc.type.toUpperCase()}: ${ioc.value}`,
          subtitle: `Added ${new Date(ioc.created_at).toLocaleDateString()}`,
          data: ioc
        });
      });

    // Add Evidence
    evidence
      .filter(ev => 
        !query || 
        ev.filename.toLowerCase().includes(query.toLowerCase()) ||
        ev.file_type.toLowerCase().includes(query.toLowerCase()) ||
        (ev.metadata && ev.metadata.toLowerCase().includes(query.toLowerCase()))
      )
      .forEach(ev => {
        items.push({
          id: `evidence-${ev.id}`,
          type: 'evidence',
          display: ev.filename,
          subtitle: `${ev.file_type} • ${formatFileSize(ev.file_size)}`,
          data: ev
        });
      });

    return items.slice(0, 10); // Limit to 10 suggestions
  }, [iocs, evidence]);

  return {
    iocs,
    evidence,
    loading,
    error,
    getMentionItems,
    refetch: () => {
      fetchIOCs();
      fetchEvidence();
    }
  };
};

// Helper function to format file sizes
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};