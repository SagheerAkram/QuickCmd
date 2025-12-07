import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import styles from '../styles/Login.module.css';

export default function Login() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            await login(username, password);
        } catch (err) {
            setError('Invalid credentials');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className={styles.container}>
            <div className={styles.card}>
                <h1>QuickCMD Login</h1>
                <p className={styles.subtitle}>Dev Mode - Default Users</p>

                <form onSubmit={handleSubmit} className={styles.form}>
                    <div className={styles.field}>
                        <label htmlFor="username">Username</label>
                        <input
                            id="username"
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="admin / approver / operator / viewer"
                            required
                        />
                    </div>

                    <div className={styles.field}>
                        <label htmlFor="password">Password</label>
                        <input
                            id="password"
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Same as username"
                            required
                        />
                    </div>

                    {error && <div className={styles.error}>{error}</div>}

                    <button type="submit" disabled={loading} className={styles.button}>
                        {loading ? 'Logging in...' : 'Login'}
                    </button>
                </form>

                <div className={styles.hint}>
                    <p><strong>Dev Mode Users:</strong></p>
                    <ul>
                        <li>admin / admin (all permissions)</li>
                        <li>approver / approver (can approve jobs)</li>
                        <li>operator / operator (can execute)</li>
                        <li>viewer / viewer (read-only)</li>
                    </ul>
                </div>
            </div>
        </div>
    );
}
