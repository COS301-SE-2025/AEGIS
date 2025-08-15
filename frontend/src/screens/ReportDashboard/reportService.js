import axios from 'axios';

// Base URL of the backend API (replace with your actual backend URL)
const API_URL = 'http://localhost:8080/api/v1';

// Generate Report for a case
export const generateReport = async (caseId, reportData) => {
  try {
    const response = await axios.post(`${API_URL}/reports/cases/${caseId}`, reportData);
    return response.data;
  } catch (error) {
    console.error('Error generating report:', error);
    throw error;
  }
};

// Get all reports for a specific case
export const getReportsByCaseID = async (caseId) => {
  try {
    const response = await axios.get(`${API_URL}/reports/cases/${caseId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching reports for case:', error);
    throw error;
  }
};

// Get all reports for a specific evidence
export const getReportsByEvidenceID = async (evidenceId) => {
  try {
    const response = await axios.get(`${API_URL}/reports/evidence/${evidenceId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching reports for evidence:', error);
    throw error;
  }
};

// Get a specific report by ID
export const getReportByID = async (reportId) => {
  try {
    const response = await axios.get(`${API_URL}/reports/${reportId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching report by ID:', error);
    throw error;
  }
};

// Update a report by ID
export const updateReport = async (reportId, updatedData) => {
  try {
    const response = await axios.put(`${API_URL}/reports/${reportId}`, updatedData);
    return response.data;
  } catch (error) {
    console.error('Error updating report:', error);
    throw error;
  }
};
