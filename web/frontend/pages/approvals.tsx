import { useEffect, useState } from 'react';
import { useApi } from '../../hooks/useAuth';

interface Approval {
    id: number;
    run_id: number;
    prompt: string;
    command: string;
    risk_level: string;
    requested_by: string;
    requested_at: string;
    required_scopes: string[];
}

export default function Approvals() {
    const [approvals, setApprovals] = useState<Approval[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedApproval, setSelectedApproval] = useState<Approval | null>(null);
    const [confirmation, setConfirmation] = useState('');
    const [note, setNote] = useState('');
    const [reason, setReason] = useState('');
    const { fetchApi } = useApi();

    useEffect(() => {
        loadApprovals();
    }, []);

    const loadApprovals = async () => {
        try {
            const data = await fetchApi('/api/v1/approvals');
            setApprovals(data.approvals || []);
        } catch (err) {
            console.error('Failed to load approvals:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleApprove = async (approval: Approval) => {
        const expectedConfirmation = `APPROVE ${approval.id}`;

        if (confirmation !== expectedConfirmation) {
            alert(`Please type exactly: ${expectedConfirmation}`);
            return;
        }

        try {
            await fetchApi(`/api/v1/approvals/${approval.id}/approve`, {
                method: 'POST',
                body: JSON.stringify({ confirmation, note }),
            });

            alert('Approval granted!');
            setSelectedApproval(null);
            setConfirmation('');
            setNote('');
            loadApprovals();
        } catch (err) {
            alert('Failed to approve: ' + err);
        }
    };

    const handleReject = async (approval: Approval) => {
        if (!reason.trim()) {
            alert('Please provide a rejection reason');
            return;
        }

        try {
            await fetchApi(`/api/v1/approvals/${approval.id}/reject`, {
                method: 'POST',
                body: JSON.stringify({ reason }),
            });

            alert('Approval rejected');
            setSelectedApproval(null);
            setReason('');
            loadApprovals();
        } catch (err) {
            alert('Failed to reject: ' + err);
        }
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div style={{ padding: '20px' }}>
            <h1>Pending Approvals</h1>

            {approvals.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '40px', color: '#6b7280' }}>
                    No pending approvals
                </div>
            ) : (
                <div>
                    {approvals.map((approval) => (
                        <div key={approval.id} style={{
                            border: '1px solid #e5e7eb',
                            borderRadius: '8px',
                            padding: '16px',
                            marginBottom: '16px',
                            backgroundColor: '#fff'
                        }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                                <div style={{ flex: 1 }}>
                                    <h3>Approval #{approval.id}</h3>
                                    <p><strong>Prompt:</strong> {approval.prompt}</p>
                                    <p><strong>Command:</strong> <code>{approval.command}</code></p>
                                    <p><strong>Risk:</strong> <span style={{
                                        color: approval.risk_level === 'high' ? '#ef4444' : '#f59e0b',
                                        fontWeight: 'bold'
                                    }}>{approval.risk_level.toUpperCase()}</span></p>
                                    <p><strong>Requested by:</strong> {approval.requested_by}</p>
                                    <p><strong>Requested at:</strong> {new Date(approval.requested_at).toLocaleString()}</p>
                                    {approval.required_scopes && approval.required_scopes.length > 0 && (
                                        <p><strong>Required scopes:</strong> {approval.required_scopes.join(', ')}</p>
                                    )}
                                </div>

                                <div style={{ display: 'flex', gap: '8px' }}>
                                    <button
                                        onClick={() => setSelectedApproval(approval)}
                                        style={{
                                            backgroundColor: '#10b981',
                                            color: 'white',
                                            padding: '8px 16px',
                                            border: 'none',
                                            borderRadius: '4px',
                                            cursor: 'pointer'
                                        }}
                                    >
                                        Approve
                                    </button>
                                    <button
                                        onClick={() => {
                                            setSelectedApproval(approval);
                                            setReason('');
                                        }}
                                        style={{
                                            backgroundColor: '#ef4444',
                                            color: 'white',
                                            padding: '8px 16px',
                                            border: 'none',
                                            borderRadius: '4px',
                                            cursor: 'pointer'
                                        }}
                                    >
                                        Reject
                                    </button>
                                </div>
                            </div>

                            {selectedApproval?.id === approval.id && (
                                <div style={{ marginTop: '16px', padding: '16px', backgroundColor: '#f9fafb', borderRadius: '4px' }}>
                                    <h4>Approve Approval #{approval.id}</h4>
                                    <p style={{ color: '#6b7280', fontSize: '14px' }}>
                                        Type <strong>APPROVE {approval.id}</strong> to confirm
                                    </p>

                                    <input
                                        type="text"
                                        placeholder={`APPROVE ${approval.id}`}
                                        value={confirmation}
                                        onChange={(e) => setConfirmation(e.target.value)}
                                        style={{ width: '100%', padding: '8px', marginBottom: '8px' }}
                                    />

                                    <textarea
                                        placeholder="Optional note..."
                                        value={note}
                                        onChange={(e) => setNote(e.target.value)}
                                        style={{ width: '100%', padding: '8px', marginBottom: '8px', minHeight: '60px' }}
                                    />

                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        <button
                                            onClick={() => handleApprove(approval)}
                                            style={{
                                                backgroundColor: '#10b981',
                                                color: 'white',
                                                padding: '8px 16px',
                                                border: 'none',
                                                borderRadius: '4px',
                                                cursor: 'pointer'
                                            }}
                                        >
                                            Confirm Approval
                                        </button>
                                        <button
                                            onClick={() => {
                                                setSelectedApproval(null);
                                                setConfirmation('');
                                                setNote('');
                                            }}
                                            style={{
                                                backgroundColor: '#6b7280',
                                                color: 'white',
                                                padding: '8px 16px',
                                                border: 'none',
                                                borderRadius: '4px',
                                                cursor: 'pointer'
                                            }}
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                </div>
                            )}

                            {selectedApproval?.id === approval.id && reason !== undefined && (
                                <div style={{ marginTop: '16px', padding: '16px', backgroundColor: '#fef2f2', borderRadius: '4px' }}>
                                    <h4>Reject Approval #{approval.id}</h4>

                                    <textarea
                                        placeholder="Rejection reason (required)..."
                                        value={reason}
                                        onChange={(e) => setReason(e.target.value)}
                                        style={{ width: '100%', padding: '8px', marginBottom: '8px', minHeight: '80px' }}
                                    />

                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        <button
                                            onClick={() => handleReject(approval)}
                                            style={{
                                                backgroundColor: '#ef4444',
                                                color: 'white',
                                                padding: '8px 16px',
                                                border: 'none',
                                                borderRadius: '4px',
                                                cursor: 'pointer'
                                            }}
                                        >
                                            Confirm Rejection
                                        </button>
                                        <button
                                            onClick={() => {
                                                setSelectedApproval(null);
                                                setReason('');
                                            }}
                                            style={{
                                                backgroundColor: '#6b7280',
                                                color: 'white',
                                                padding: '8px 16px',
                                                border: 'none',
                                                borderRadius: '4px',
                                                cursor: 'pointer'
                                            }}
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
