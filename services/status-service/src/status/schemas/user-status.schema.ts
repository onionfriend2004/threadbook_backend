export interface UserStatus {
  userId: number;
  username: string;
  customStatus: string | null;
  isPrivate: boolean;
  isOnline: boolean;
  lastSeen: number | null;
}
