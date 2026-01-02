import { useState, useEffect, useRef } from 'react';

const LogLevel = {
  debug: { color: 'text-gray-400', bg: 'bg-gray-700', label: 'DEBUG' },
  info: { color: 'text-blue-400', bg: 'bg-blue-900', label: 'INFO' },
  warning: { color: 'text-yellow-400', bg: 'bg-yellow-900', label: 'WARN' },
  error: { color: 'text-red-400', bg: 'bg-red-900', label: 'ERROR' },
};

const StepBadge = ({ step }) => {
  if (!step) return null;

  const stepStyles = {
    init: 'bg-purple-100 text-purple-800',
    snapshot: 'bg-indigo-100 text-indigo-800',
    generate_code: 'bg-blue-100 text-blue-800',
    write_files: 'bg-cyan-100 text-cyan-800',
    docker_build: 'bg-orange-100 text-orange-800',
    docker_start: 'bg-green-100 text-green-800',
    complete: 'bg-emerald-100 text-emerald-800',
  };

  const style = stepStyles[step] || 'bg-gray-100 text-gray-800';
  const label = step.replace(/_/g, ' ').toUpperCase();

  return (
    <span className={`px-2 py-0.5 text-xs font-medium rounded ${style}`}>
      {label}
    </span>
  );
};

const LogEntry = ({ log, showTimestamp = true }) => {
  const levelStyle = LogLevel[log.level] || LogLevel.info;
  const timestamp = new Date(log.created_at).toLocaleTimeString();

  return (
    <div className="flex items-start gap-2 py-1 px-2 hover:bg-gray-800 font-mono text-sm">
      {showTimestamp && (
        <span className="text-gray-500 shrink-0">{timestamp}</span>
      )}
      <span className={`px-1.5 py-0.5 text-xs font-semibold rounded ${levelStyle.bg} ${levelStyle.color}`}>
        {levelStyle.label}
      </span>
      {log.step && <StepBadge step={log.step} />}
      <span className="text-gray-200 break-all">{log.message}</span>
    </div>
  );
};

const ContainerLogEntry = ({ log }) => {
  const timestamp = log.timestamp ? new Date(log.timestamp).toLocaleTimeString() : '';
  const streamColor = log.stream === 'stderr' ? 'text-red-400' : 'text-green-400';

  return (
    <div className="flex items-start gap-2 py-0.5 px-2 hover:bg-gray-800 font-mono text-sm">
      <span className="text-gray-500 shrink-0">{timestamp}</span>
      <span className={`shrink-0 ${streamColor}`}>
        [{log.stream || 'stdout'}]
      </span>
      <span className="text-gray-200 break-all">{log.message}</span>
    </div>
  );
};

export default function LogViewer({
  logs = [],
  isStreaming = false,
  onClose,
  title = 'Logs',
  type = 'deployment', // 'deployment' or 'container'
  maxHeight = '400px',
  autoScroll = true
}) {
  const [filter, setFilter] = useState('all');
  const [searchTerm, setSearchTerm] = useState('');
  const logsEndRef = useRef(null);
  const containerRef = useRef(null);
  const [shouldAutoScroll, setShouldAutoScroll] = useState(autoScroll);

  // Filter logs
  const filteredLogs = logs.filter(log => {
    if (filter !== 'all' && log.level !== filter) return false;
    if (searchTerm && !log.message.toLowerCase().includes(searchTerm.toLowerCase())) return false;
    return true;
  });

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (shouldAutoScroll && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, shouldAutoScroll]);

  // Detect manual scroll
  const handleScroll = () => {
    if (containerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
      const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;
      setShouldAutoScroll(isAtBottom);
    }
  };

  return (
    <div className="bg-gray-900 rounded-lg border border-gray-700 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center gap-3">
          <h3 className="text-white font-medium">{title}</h3>
          {isStreaming && (
            <span className="flex items-center gap-1.5 text-green-400 text-sm">
              <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
              Live
            </span>
          )}
          <span className="text-gray-400 text-sm">
            {filteredLogs.length} {filteredLogs.length === 1 ? 'entry' : 'entries'}
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* Search */}
          <input
            type="text"
            placeholder="Search logs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="px-3 py-1.5 text-sm bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />

          {/* Filter by level */}
          {type === 'deployment' && (
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="px-3 py-1.5 text-sm bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="all">All Levels</option>
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warning">Warning</option>
              <option value="error">Error</option>
            </select>
          )}

          {/* Auto-scroll toggle */}
          <button
            onClick={() => setShouldAutoScroll(!shouldAutoScroll)}
            className={`p-2 rounded ${shouldAutoScroll ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-400'}`}
            title={shouldAutoScroll ? 'Auto-scroll enabled' : 'Auto-scroll disabled'}
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
            </svg>
          </button>

          {onClose && (
            <button
              onClick={onClose}
              className="p-2 text-gray-400 hover:text-white"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>
      </div>

      {/* Logs container */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="overflow-y-auto bg-gray-900"
        style={{ maxHeight }}
      >
        {filteredLogs.length === 0 ? (
          <div className="flex items-center justify-center py-8 text-gray-500">
            {logs.length === 0 ? 'No logs available' : 'No logs match your filter'}
          </div>
        ) : (
          <div className="py-2">
            {filteredLogs.map((log, index) => (
              type === 'container' ? (
                <ContainerLogEntry key={log.id || index} log={log} />
              ) : (
                <LogEntry key={log.id || index} log={log} />
              )
            ))}
            <div ref={logsEndRef} />
          </div>
        )}
      </div>

      {/* Footer with stats */}
      {type === 'deployment' && logs.length > 0 && (
        <div className="flex items-center gap-4 px-4 py-2 bg-gray-800 border-t border-gray-700 text-xs text-gray-400">
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 bg-blue-400 rounded-full"></span>
            Info: {logs.filter(l => l.level === 'info').length}
          </span>
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 bg-yellow-400 rounded-full"></span>
            Warning: {logs.filter(l => l.level === 'warning').length}
          </span>
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 bg-red-400 rounded-full"></span>
            Error: {logs.filter(l => l.level === 'error').length}
          </span>
        </div>
      )}
    </div>
  );
}
