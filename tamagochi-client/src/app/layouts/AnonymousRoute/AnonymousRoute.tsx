import { useAppSelector } from '@/store/hooks';
import { Navigate, Outlet } from 'react-router';

export const AnonymousRoute = () => {
  const userStatus = useAppSelector((state) => state.auth.authStatus);

  if (userStatus === 'authenticated') {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
};
