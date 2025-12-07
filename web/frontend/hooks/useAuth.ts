import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';

interface User {
    username: string;
    roles: string[];
}

interface AuthState {
    user: User | null;
    token: string | null;
    loading: boolean;
}

export function useAuth() {
    const [authState, setAuthState] = useState<AuthState>({
        user: null,
        token: null,
        loading: true,
    });
    const router = useRouter();

    useEffect(() => {
        // Check for token in localStorage
        const token = localStorage.getItem('quickcmd_token');
        if (token) {
            // In production, validate token with backend
            setAuthState({
                user: { username: 'user', roles: [] }, // Parse from JWT
                token,
                loading: false,
            });
        } else {
            setAuthState({ user: null, token: null, loading: false });
        }
    }, []);

    const login = async (username: string, password: string) => {
        const response = await fetch('/api/v1/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
        });

        if (!response.ok) {
            throw new Error('Login failed');
        }

        const data = await response.json();
        localStorage.setItem('quickcmd_token', data.token);

        setAuthState({
            user: { username, roles: [] },
            token: data.token,
            loading: false,
        });

        router.push('/history');
    };

    const logout = () => {
        localStorage.removeItem('quickcmd_token');
        setAuthState({ user: null, token: null, loading: false });
        router.push('/login');
    };

    return { ...authState, login, logout };
}

export function useApi() {
    const { token } = useAuth();

    const fetchApi = async (url: string, options: RequestInit = {}) => {
        const headers = {
            'Content-Type': 'application/json',
            ...(token && { Authorization: `Bearer ${token}` }),
            ...options.headers,
        };

        const response = await fetch(url, { ...options, headers });

        if (response.status === 401) {
            // Redirect to login
            window.location.href = '/login';
            throw new Error('Unauthorized');
        }

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Request failed');
        }

        return response.json();
    };

    return { fetchApi };
}
