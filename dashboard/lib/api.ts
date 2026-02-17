const API_BASE = "http://localhost:8080";

export interface ExperimentResponse {
    user_id: string;
    segments: string[];
    features: {
        show_pizza_tile?: boolean;
        home_banner?: string;
        discount_pct?: number;
    };
}

export const api = {
    getUsers: () => fetch(`${API_BASE}/users`).then(res => res.json()),

    createUser: (userId: string) =>
        fetch(`${API_BASE}/users`, {
            method: 'POST',
            body: JSON.stringify({ user_id: userId }),
        }),

    placeOrders: (userId: string, count: number, instant: boolean) =>
        fetch(`${API_BASE}/place-order`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: userId,
                count: count,
                instant_sync: instant // The worker will look for this
            }),
        }),

    getExperiments: (userId: string): Promise<ExperimentResponse> =>
        fetch(`${API_BASE}/experiments?userId=${userId}`).then(res => res.json()),
};