import { render, screen } from '@testing-library/react';
import { PasswordField } from '../PasswordField';
import userEvent from '@testing-library/user-event';

describe('PasswordField', () => {
  it('should render component with hidden password', () => {
    render(<PasswordField label="Password" />);
    const input = screen.getByLabelText(/password/i);

    expect(input).toHaveAttribute('type', 'password');
  });

  it('should show password after click on the eye button', async () => {
    render(<PasswordField label="Password" />);

    const input = screen.getByLabelText(/password/i);

    const toggleButton = screen.getByRole('button', { name: 'Показать пароль' });

    await userEvent.click(toggleButton);

    expect(input).toHaveAttribute('type', 'text');
    expect(screen.getByRole('button', { name: 'Скрыть пароль' })).toBeInTheDocument();
  });
});
