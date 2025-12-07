import { useEffect, useState } from 'react';
import { useApi } from '../hooks/useAuth';
import Link from 'next/link';

interface RunRecord {
    id: number;
    timestamp: string;
    prompt: string;
    selected_command: string;
    risk_level: string;
    executed: boolean;
    exit_code?: number;
}

export default function History() {
    const [records, setRecords] = useState<RunRecord[]>([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState('');
    const { fetchApi } = useApi();

    useEffect(() => {
        loadHistory();
    }, [filter]);

    const loadHistory = async () => {
        try {
            const data = await fetchApi(`/api/v1/history?limit=50&filter=${filter}`);
            setRecords(data.records || []);
        } catch (err) {
            console.error('Failed to load history:', err);
        } finally {
            setLoading(false);
        }
    };

    const getRiskBadge = (risk: string) => {
        const colors = {
            safe: '#10b981',
            medium: '#f59e0b',
            high: '#ef4444',
        };
        return (
            <span style={{
                backgroundColor: colors[risk as keyof typeof colors] || '#6b7280',
                color: 'white',
                padding: '2px 8px',
                borderRadius: '4px',
                fontSize: '12px',
                fontWeight: 'bold'
            }}>
                {risk.toUpperCase()}
            </span>
        );
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div style={{ padding: '20px' }}>
            <h1>Execution History</h1>

            <div style={{ marginBottom: '20px' }}>
                <input
                    type="text"
                    placeholder="Filter by keyword..."
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    style={{ padding: '8px', width: '300px', marginRight: '10px' }}
                />
                <button onClick={loadHistory}>Refresh</button>
            </div>

            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                    <tr style={{ backgroundColor: '#f3f4f6', textAlign: 'left' }}>
                        <th style={{ padding: '12px' }}>ID</th>
                        <th style={{ padding: '12px' }}>Timestamp</th>
                        <th style={{ padding: '12px' }}>Prompt</th>
                        <th style={{ padding: '12px' }}>Command</th>
                        <th style={{ padding: '12px' }}>Risk</th>
                        <th style={{ padding: '12px' }}>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {records.map((record) => (
                        <tr key={record.id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                            <td style={{ padding: '12px' }}>
                                <Link href={`/run/${record.id}`} style={{ color: '#3b82f6' }}>
                                    #{record.id}
                                </Link>
                            </td>
                            <td style={{ padding: '12px', fontSize: '14px' }}>
                                {new Date(record.timestamp).toLocaleString()}
                            </td>
                            <td style={{ padding: '12px' }}>{record.prompt}</td>
                            <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '13px' }}>
                                {record.selected_command.substring(0, 50)}...
                            </td>
                            <td style={{ padding: '12px' }}>{getRiskBadge(record.risk_level)}</td>
                            <td style={{ padding: '12px' }}>
                                {record.executed ? (
                                    <span style={{ color: record.exit_code === 0 ? '#10b981' : '#ef4444' }}>
                                        {record.exit_code === 0 ? '✓ Success' : '✗ Failed'}
                                    </span>
                                ) : (
                                    <span style={{ color: '#6b7280' }}>Dry-run</span>
                                )}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {records.length === 0 && (
                <div style={{ textAlign: 'center', padding: '40px', color: '#6b7280' }}>
                    No execution history found
                </div>
            )}
        </div>
    );
}
