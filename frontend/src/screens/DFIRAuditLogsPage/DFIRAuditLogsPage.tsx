import { useState, useEffect } from 'react';
import { 
  Shield, 
  User, 
  FileText, 
  Database, 
  Activity, 
  Clock, 
  Search, 
  Filter, 
  Download,
  AlertTriangle,
  CheckCircle,
  XCircle,
  RefreshCw,
  ChevronDown,
  ChevronRight,
    ArrowLeft
} from 'lucide-react';
import axios from 'axios';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { useNavigate } from 'react-router-dom';

interface AuditLog {
  ID: string;  // Changed from 'id' to 'ID'
  Timestamp: string;  // Changed from 'timestamp' to 'Timestamp'
  Action: string;  // Changed from 'action' to 'Action'
  Actor: {  // Changed from 'actor' to 'Actor'
    ID: string;  // Changed from 'id' to 'ID'
    Email: string;  // Changed from 'email' to 'Email'
    Role: string;  // Changed from 'role' to 'Role'
    IP: string;  // Changed from 'ip' to 'IP'
    UA: string;  // Changed from 'ua' to 'UA'
  };
  Target: {  // Changed from 'target' to 'Target'
    Type: string;  // Changed from 'type' to 'Type'
    ID: string;  // Changed from 'id' to 'ID'
    Extra?: any;  // Changed from 'extra' to 'Extra'
  };
  Service: string;  // Changed from 'service' to 'Service'
  Status: 'SUCCESS' | 'FAILED';  // Changed from 'status' to 'Status'
  Description: string;  // Changed from 'description' to 'Description'
  Metadata?: Record<string, string>;  // Changed from 'metadata' to 'Metadata'
}
interface AuditLogFilter {
  status: 'ALL' | 'SUCCESS' | 'FAILED';
  action: string;
  service: string;
  limit: number;
  dateFrom?: string;
  dateTo?: string;
  actorId?: string;
}

export const DFIRAuditLogsPage = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<AuditLogFilter>({
    status: 'ALL',
    action: '',
    service: '',
    limit: 100
  });
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('ALL');
  const [expandedLogs, setExpandedLogs] = useState<Set<string>>(new Set());
  const navigate = useNavigate();

// Add this function to dynamically create categories based on actual data
const getDynamicCategories = (logs: AuditLog[]) => {
  const uniqueActions = [...new Set(logs.map(log => log.Action))];
  
  return [
    { 
      id: 'ALL', 
      name: 'All Logs', 
      icon: Activity, 
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
      borderColor: 'border-blue-500/20'
    },
    { 
      id: 'USER', 
      name: 'User Activities', 
      icon: User, 
      color: 'text-green-500',
      bgColor: 'bg-green-500/10',
      borderColor: 'border-green-500/20',
      actions: uniqueActions.filter(action => 
        action.includes('USER') || action.includes('LOGIN') || action.includes('PASSWORD')
      )
    },
    { 
      id: 'CASE', 
      name: 'Case Management', 
      icon: FileText, 
      color: 'text-orange-500',
      bgColor: 'bg-orange-500/10',
      borderColor: 'border-orange-500/20',
       actions: ['CASE_CREATED', 'CASE_UPDATED', 'CASE_DELETED', 'CASE_ASSIGNED', 'LIST_FILTERED_CASES', 'LIST_ACTIVE_CASES', 'LIST_CLOSED_CASES', 'LIST_ARCHIVED_CASES']

    },
    { 
      id: 'EVIDENCE', 
      name: 'Evidence Management', 
      icon: Database, 
      color: 'text-purple-500',
      bgColor: 'bg-purple-500/10',
      borderColor: 'border-purple-500/20',
      actions: uniqueActions.filter(action => 
        action.includes('EVIDENCE')
      )
    }
  ];
};

// Then use it in your component:
const logCategories = getDynamicCategories(logs);

  useEffect(() => {
    fetchAuditLogs();
  }, [filter.status, filter.action, filter.service, filter.limit]); // Remove selectedCategory

const fetchAuditLogs = async () => {
  try {
    setLoading(true);
    const token = sessionStorage.getItem('authToken');
    
    if (!token) {
      toast.error('Authentication required');
      return;
    }

    const queryParams = new URLSearchParams();
    if (filter.status !== 'ALL') queryParams.append('status', filter.status);
    if (filter.action) queryParams.append('action', filter.action);
    if (filter.service) queryParams.append('service', filter.service);
    queryParams.append('limit', filter.limit.toString());
    if (filter.dateFrom) queryParams.append('dateFrom', filter.dateFrom);
    if (filter.dateTo) queryParams.append('dateTo', filter.dateTo);

    const url = `https://localhost/api/v1/audit-logs?${queryParams.toString()}`;
    console.log('🔍 Fetching from URL:', url);

    const response = await axios.get<{
      data?: any;
      logs?: AuditLog[];
      [key: string]: any;
    }>(url, {
      headers: { Authorization: `Bearer ${token}` }
    });

    console.log('✅ Response status:', response.status);
    console.log('📦 Raw response.data:', response.data);
    if (response.data && typeof response.data === 'object') {
      console.log('🔍 Response.data keys:', Object.keys(response.data));
    } else {
      console.log('🔍 Response.data is not an object:', response.data);
    }
    
    // Check the actual data structure
    console.log('🔍 Response.data.data:', response.data.data);
    console.log('🔍 Type of response.data.data:', typeof response.data.data);

    let logsData: AuditLog[] = [];
    
    // Handle the new structure: { success, message, data }
    if (response.data && response.data.data) {
      console.log('📋 Found data in response.data.data');
      
      // Check if data.data is an array
      if (Array.isArray(response.data.data)) {
        console.log('📋 response.data.data is array:', response.data.data.length);
        logsData = response.data.data;
      }
      // Check if data.data has a logs property
      else if (response.data.data.logs && Array.isArray(response.data.data.logs)) {
        console.log('📋 response.data.data.logs is array:', response.data.data.logs.length);
        logsData = response.data.data.logs;
      }
      // Check if data.data has other properties that might contain logs
      else if (typeof response.data.data === 'object') {
        console.log('📋 response.data.data is object, keys:', Object.keys(response.data.data));
        
        // Check for common property names
        if (response.data.data.logs) {
          logsData = response.data.data.logs;
        } else if (response.data.data.items) {
          logsData = response.data.data.items;
        } else if (response.data.data.results) {
          logsData = response.data.data.results;
        } else {
          console.log('❌ Could not find logs in data.data object');
        }
      }
    }
    // Fallback to original structure checking
    else if (Array.isArray(response.data)) {
      console.log('📋 Data is direct array');
      logsData = response.data;
    } else if (response.data.logs && Array.isArray(response.data.logs)) {
      console.log('📋 Data has logs property');
      logsData = response.data.logs;
    } else {
      console.log('❌ Unexpected data structure');
      console.log('📋 Available keys:', Object.keys(response.data));
    }

    console.log('🎯 Final logs data to set:', logsData);
    console.log('🎯 Final logs data length:', logsData.length);
    
    if (logsData.length > 0) {
      console.log('🔍 First log item structure:', logsData[0]);
    }
    
    setLogs(logsData);

  } catch (error: any) {
    console.error('❌ Failed to fetch audit logs:', error);
    console.error('❌ Error response:', error.response?.data);
    toast.error('Failed to fetch audit logs');
    setLogs([]);
  } finally {
    setLoading(false);
  }
};

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'SUCCESS':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'FAILED':
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />;
    }
  };

  const getActionColor = (action: string) => {
    if (action.includes('DELETE') || action.includes('FAILED')) return 'text-red-400';
    if (action.includes('CREATE') || action.includes('SUCCESS')) return 'text-green-400';
    if (action.includes('UPDATE') || action.includes('MODIFY')) return 'text-yellow-400';
    return 'text-blue-400';
  };

// Updated the filtering logic to be more flexible:
const matchesCategory = (log: AuditLog, categoryActions?: string[]) => {
  if (!categoryActions) return false;
  
  return categoryActions.some(action => {
    // Exact match
    if (log.Action === action) return true;
    
    // Partial match (log action contains category action)
    if (log.Action.includes(action)) return true;
    
    // Reverse partial match (category action contains log action)
    if (action.includes(log.Action)) return true;
    
    // Case-insensitive match
    if (log.Action.toLowerCase() === action.toLowerCase()) return true;
    
    return false;
  });
};

  // Update the search filtering to be more comprehensive
  const filteredLogs = logs.filter(log => {
    // Filter by category first
    if (selectedCategory !== 'ALL') {
      const category = logCategories.find(cat => cat.id === selectedCategory);
      if (!matchesCategory(log, category?.actions)) {
        return false;
      }
    }

    // Apply date filters
    if (filter.dateFrom) {
      const logDate = new Date(log.Timestamp).toISOString().split('T')[0];
      if (logDate < filter.dateFrom) {
        return false;
      }
    }

    if (filter.dateTo) {
      const logDate = new Date(log.Timestamp).toISOString().split('T')[0];
      if (logDate > filter.dateTo) {
        return false;
      }
    }

    // Filter by search term - make it more comprehensive
    if (searchTerm && searchTerm.trim() !== '') {
      const searchLower = searchTerm.toLowerCase().trim();
      const searchFields = [
        log.Action,
        log.Actor.Email,
        log.Actor.Role,
        log.Description,
        log.Service,
        log.Status,
        log.Target.Type,
        log.Actor.IP,
        // Include metadata in search
        log.Metadata ? Object.values(log.Metadata).join(' ') : '',
      ].filter(Boolean); // Remove any null/undefined values

      return searchFields.some(field => 
        field.toString().toLowerCase().includes(searchLower)
      );
    }

    return true;
  });
  const toggleLogExpansion = (logId: string) => {
    const newExpanded = new Set(expandedLogs);
    if (newExpanded.has(logId)) {
      newExpanded.delete(logId);
    } else {
      newExpanded.add(logId);
    }
    setExpandedLogs(newExpanded);
  };

  const exportLogs = async () => {
    try {
      const token = sessionStorage.getItem('authToken');
      const response = await axios.get(
        'https://localhost/api/v1/audit-logs/export',
        {
          headers: { Authorization: `Bearer ${token}` },
          responseType: 'blob'
        }
      );

      const blob = new Blob([response.data as BlobPart], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `audit_logs_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast.success('Audit logs exported successfully');
    } catch (error) {
      toast.error('Failed to export audit logs');
    }
  };

  console.log('Current logs state:', logs);
  console.log('Filtered logs:', filteredLogs);
  console.log('Selected category:', selectedCategory);
  console.log('Search term:', searchTerm);
  console.log('Loading state:', loading);

  return (
    <div className="min-h-screen bg-background text-foreground p-6">
      <ToastContainer
        position="top-right"
        theme="dark"
        toastClassName="bg-card text-foreground border border-border"
      />

      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-3">
              <Shield className="w-8 h-8 text-red-500" />
              DFIR Audit Dashboard
            </h1>
            <p className="text-muted-foreground mt-2">
              Comprehensive system activity monitoring and forensic analysis
            </p>
          </div>
          <div className="flex gap-3">
                <button
                    onClick={() => navigate(-1)}
                    className="flex items-center gap-2 text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors border border-border hover:bg-muted"
                  >
                    <ArrowLeft className="w-4 h-4" />
                    Back
                  </button>            
            <button
              onClick={fetchAuditLogs}
              className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
            >
              <RefreshCw className="w-4 h-4" />
              Refresh
            </button>
            <button
              onClick={exportLogs}
              className="flex items-center gap-2 px-4 py-2 bg-secondary text-secondary-foreground rounded-lg hover:bg-secondary/90 transition-colors"
            >
              <Download className="w-4 h-4" />
              Export
            </button>
          </div>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-muted-foreground">Total Events</p>
              <p className="text-2xl font-bold">{filteredLogs.length}</p> {/* Changed from logs.length */}
              {filteredLogs.length !== logs.length && (
                <p className="text-xs text-muted-foreground">({logs.length} total)</p>
              )}
            </div>
            <Activity className="w-8 h-8 text-blue-500" />
          </div>
        </div>
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-muted-foreground">Success Rate</p>
              <p className="text-2xl font-bold text-green-500">
                {filteredLogs.length > 0 ? Math.round((filteredLogs.filter(l => l.Status === 'SUCCESS').length / filteredLogs.length) * 100) : 0}%
              </p>
            </div>
            <CheckCircle className="w-8 h-8 text-green-500" />
          </div>
        </div>
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-muted-foreground">Failed Events</p>
              <p className="text-2xl font-bold text-red-500">
                {filteredLogs.filter(l => l.Status === 'FAILED').length}
              </p>
            </div>
            <XCircle className="w-8 h-8 text-red-500" />
          </div>
        </div>
     
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Category Sidebar */}
        <div className="lg:col-span-1">
          <div className="bg-card border border-border rounded-lg p-6">
            <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <Filter className="w-5 h-5" />
              Log Categories
            </h3>
            <div className="space-y-2">
              {logCategories.map(category => {
                const Icon = category.icon;
                const isSelected = selectedCategory === category.id;
                const count = category.id === 'ALL' 
                  ? logs.length 
                  : logs.filter(log => matchesCategory(log, category.actions)).length;

                return (
                  <button
                    key={category.id}
                    onClick={() => setSelectedCategory(category.id)}
                    className={`w-full flex items-center justify-between p-3 rounded-lg border transition-all ${
                      isSelected 
                        ? `${category.bgColor} ${category.borderColor} ${category.color}` 
                        : 'border-border hover:bg-muted'
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      <Icon className={`w-5 h-5 ${isSelected ? category.color : 'text-muted-foreground'}`} />
                      <span className={`font-medium ${isSelected ? category.color : 'text-foreground'}`}>
                        {category.name}
                      </span>
                    </div>
                    <span className={`text-sm px-2 py-1 rounded-full ${
                      isSelected ? category.bgColor : 'bg-muted text-muted-foreground'
                    }`}>
                      {count}
                    </span>
                  </button>
                );
              })}
            </div>

            {/* Filters */}
            <div className="mt-6 pt-6 border-t border-border">
              <h4 className="font-semibold mb-3">Filters</h4>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm text-muted-foreground mb-1">Status</label>
                  <select
                    value={filter.status}
                    onChange={(e) => setFilter(prev => ({ ...prev, status: e.target.value as any }))}
                    className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground"
                  >
                    <option value="ALL">All Status</option>
                    <option value="SUCCESS">Success</option>
                    <option value="FAILED">Failed</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm text-muted-foreground mb-1">Service</label>
                  <select
                    value={filter.service}
                    onChange={(e) => setFilter(prev => ({ ...prev, service: e.target.value }))}
                    className="w-full px-3 py-2 bg-input border border-border rounded-lg text-foreground"
                  >
                    <option value="">All Services</option>
                    <option value="evidence">Evidence</option>
                    <option value="case">Case</option>
                    <option value="user">User</option>
                    <option value="auth">Authentication</option>
                    <option value="report">Report</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="lg:col-span-3">
          {/* Search and Controls */}
          <div className="bg-card border border-border rounded-lg p-6 mb-6">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1 relative">
                <Search className="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="Search logs by action, user, email, description, IP..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 bg-input border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
              <div className="flex gap-2 items-center">
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-muted-foreground">From Date</label>
                  <input
                    type="date"
                    value={filter.dateFrom || ''}
                    max={new Date().toISOString().split('T')[0]} // Prevent future dates
                    onChange={(e) => {
                      const selectedDate = e.target.value;
                      const today = new Date().toISOString().split('T')[0];
                      
                      // Validate date is not in the future
                      if (selectedDate > today) {
                        toast.error('Cannot select future dates');
                        return;
                      }
                      
                      // Validate that from date is not after to date
                      if (filter.dateTo && selectedDate > filter.dateTo) {
                        toast.error('From date cannot be after To date');
                        return;
                      }
                      
                      setFilter(prev => ({ ...prev, dateFrom: selectedDate }));
                    }}
                    className="px-3 py-2 bg-input border border-border rounded-lg text-foreground text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-muted-foreground">To Date</label>
                  <input
                    type="date"
                    value={filter.dateTo || ''}
                    max={new Date().toISOString().split('T')[0]} // Prevent future dates
                    min={filter.dateFrom || undefined} // Ensure to date is after from date
                    onChange={(e) => {
                      const selectedDate = e.target.value;
                      const today = new Date().toISOString().split('T')[0];
                      
                      // Validate date is not in the future
                      if (selectedDate > today) {
                        toast.error('Cannot select future dates');
                        return;
                      }
                      
                      // Validate that to date is not before from date
                      if (filter.dateFrom && selectedDate < filter.dateFrom) {
                        toast.error('To date cannot be before From date');
                        return;
                      }
                      
                      setFilter(prev => ({ ...prev, dateTo: selectedDate }));
                    }}
                    className="px-3 py-2 bg-input border border-border rounded-lg text-foreground text-sm"
                  />
                </div>
                {/* Clear date filters button */}
                {(filter.dateFrom || filter.dateTo) && (
                  <button
                    onClick={() => setFilter(prev => ({ ...prev, dateFrom: undefined, dateTo: undefined }))}
                    className="px-3 py-2 bg-muted text-muted-foreground rounded-lg hover:bg-muted/80 transition-colors text-sm mt-5"
                    title="Clear date filters"
                  >
                    Clear
                  </button>
                )}
              </div>
            </div>
            
            {/* Show active filters */}
            {(searchTerm || filter.dateFrom || filter.dateTo || selectedCategory !== 'ALL') && (
              <div className="mt-4 pt-4 border-t border-border">
                <div className="flex flex-wrap gap-2 items-center">
                  <span className="text-sm text-muted-foreground">Active filters:</span>
                  
                  {searchTerm && (
                    <span className="bg-primary/10 text-primary px-2 py-1 rounded-full text-xs flex items-center gap-1">
                      Search: "{searchTerm}"
                      <button 
                        onClick={() => setSearchTerm('')}
                        className="hover:bg-primary/20 rounded-full p-0.5"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  
                  {selectedCategory !== 'ALL' && (
                    <span className="bg-orange-500/10 text-orange-500 px-2 py-1 rounded-full text-xs flex items-center gap-1">
                      Category: {logCategories.find(c => c.id === selectedCategory)?.name}
                      <button 
                        onClick={() => setSelectedCategory('ALL')}
                        className="hover:bg-orange-500/20 rounded-full p-0.5"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  
                  {filter.dateFrom && (
                    <span className="bg-blue-500/10 text-blue-500 px-2 py-1 rounded-full text-xs flex items-center gap-1">
                      From: {filter.dateFrom}
                      <button 
                        onClick={() => setFilter(prev => ({ ...prev, dateFrom: undefined }))}
                        className="hover:bg-blue-500/20 rounded-full p-0.5"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  
                  {filter.dateTo && (
                    <span className="bg-blue-500/10 text-blue-500 px-2 py-1 rounded-full text-xs flex items-center gap-1">
                      To: {filter.dateTo}
                      <button 
                        onClick={() => setFilter(prev => ({ ...prev, dateTo: undefined }))}
                        className="hover:bg-blue-500/20 rounded-full p-0.5"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  
                  <button
                    onClick={() => {
                      setSearchTerm('');
                      setSelectedCategory('ALL');
                      setFilter(prev => ({ ...prev, dateFrom: undefined, dateTo: undefined }));
                    }}
                    className="text-xs text-muted-foreground hover:text-foreground underline"
                  >
                    Clear all filters
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Logs List */}
          <div className="bg-card border border-border rounded-lg">
            <div className="p-6 border-b border-border">
              <h3 className="text-lg font-semibold flex items-center gap-2">
                <Clock className="w-5 h-5" />
                Audit Trail ({filteredLogs.length} events)
              </h3>
            </div>
            <div className="max-h-[600px] overflow-y-auto">
              {loading ? (
                <div className="p-8 text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
                  <p className="text-muted-foreground">Loading audit logs...</p>
                </div>
              ) : filteredLogs.length === 0 ? (
                <div className="p-8 text-center">
                  <AlertTriangle className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No audit logs found matching your criteria</p>
                </div>
              ) : (
                <div className="divide-y divide-border">
                  {filteredLogs.map((log) => {
                    const isExpanded = expandedLogs.has(log.ID);  // Changed from log.id
                    return (
                      <div key={log.ID} className="p-4 hover:bg-muted/50 transition-colors">  {/* Changed from log.id */}
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center gap-3 mb-2">
                              {getStatusIcon(log.Status)}  {/* Changed from log.status */}
                              <span className={`font-mono text-sm px-2 py-1 rounded ${getActionColor(log.Action)} bg-muted`}>
                                {log.Action}  {/* Changed from log.action */}
                              </span>
                              <span className="text-sm text-muted-foreground">
                                {log.Service}  {/* Changed from log.service */}
                              </span>
                              <span className="text-sm text-muted-foreground">
                                {new Date(log.Timestamp).toLocaleString()}  {/* Changed from log.timestamp */}
                              </span>
                            </div>
                            <div className="flex items-center gap-2 mb-2">
                              <User className="w-4 h-4 text-muted-foreground" />
                              <span className="text-sm">{log.Actor.Email}</span>  {/* Changed from log.actor.email */}
                            </div>
                            <p className="text-sm text-muted-foreground">{log.Description}</p>  {/* Changed from log.description */}
                          </div>
                          <button
                            onClick={() => toggleLogExpansion(log.ID)}  // Changed from log.id
                            className="p-1 hover:bg-muted rounded"
                          >
                            {isExpanded ? 
                              <ChevronDown className="w-4 h-4" /> : 
                              <ChevronRight className="w-4 h-4" />
                            }
                          </button>
                        </div>
                        
                        {isExpanded && (
                          <div className="mt-4 p-4 bg-muted/30 rounded-lg border border-border">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                              <div>
                                <h5 className="font-semibold mb-2">Actor Details</h5>
                                <div className="space-y-1 text-muted-foreground">
                                  <p><span className="font-medium">ID:</span> {log.Actor.ID}</p>  {/* Changed from log.actor.id */}
                                  <p><span className="font-medium">Role:</span> {log.Actor.Role}</p>  {/* Changed from log.actor.role */}
                                  <p><span className="font-medium">User Agent:</span> {log.Actor.UA}</p>  {/* Changed from log.actor.ua */}
                                </div>
                              </div>
                              <div>
                                <h5 className="font-semibold mb-2">Target Details</h5>
                                <div className="space-y-1 text-muted-foreground">
                                  <p><span className="font-medium">Type:</span> {log.Target.Type}</p>  {/* Changed from log.target.type */}
                                  <p><span className="font-medium">ID:</span> {log.Target.ID}</p>  {/* Changed from log.target.id */}
                                  {log.Target.Extra && (  // Changed from log.target.extra
                                    <p><span className="font-medium">Extra:</span> {JSON.stringify(log.Target.Extra)}</p>
                                  )}
                                </div>
                              </div>
                              {log.Metadata && Object.keys(log.Metadata).length > 0 && (  // Changed from log.metadata
                                <div className="md:col-span-2">
                                  <h5 className="font-semibold mb-2">Metadata</h5>
                                  <div className="bg-background/50 rounded p-2 font-mono text-xs">
                                    <pre>{JSON.stringify(log.Metadata, null, 2)}</pre>  {/* Changed from log.metadata */}
                                  </div>
                                </div>
                              )}
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

