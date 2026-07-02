import { Outlet, NavLink } from 'react-router';

export default function CaptureLayOut() {
  const items = [
    { path: 'control', label: 'Capture Control' },
    { path: 'sessions', label: 'Sessions' },
  ];

  return (
    <div className="p-6">
      <div className="mb-4 flex gap-4">
        {items.map(({ path, label }) => (
          <NavLink
            key={path}
            to={path}
            className={({ isActive }) =>
              `px-4 py-2 rounded-md text-sm font-medium ${
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              }`
            }
          >
            {label}
          </NavLink>
        ))}
      </div>
      <Outlet />
    </div>
  );
}
