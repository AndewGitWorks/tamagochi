import { useAppSelector } from '@/store/hooks';
import { Navigate, Outlet } from 'react-router';

export const ProtectedRoute = () => {
  const userStatus = useAppSelector((state) => state.auth.authStatus);

  if (userStatus === 'unauthenticated') {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
};
