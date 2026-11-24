import { render, screen } from '@testing-library/react';
import { TextField } from '../TextField';
import userEvent from '@testing-library/user-event';

describe('TextField', () => {
  it('should support typing text', async () => {
    render(<TextField label="Name" />);
    const input = screen.getByRole('textbox');

    await userEvent.type(input, 'Steven Le');
    expect(input).toHaveValue('Steven Le');
  });

  it('should display error message when error prop is provided', () => {
    render(<TextField label="Error" error="Invalid email address" />);

    expect(screen.getByText('Invalid email address')).toBeInTheDocument();
  });

  it('should render label and input with correct connection without explicit id', async () => {
    render(<TextField label="Email" />);

    const label = screen.getByText('Email:');
    const input = screen.getByRole('textbox');

    await userEvent.click(label);
    expect(input).toHaveFocus();
  });

  it('should be disabled when disabled props is provided', async () => {
    render(<TextField label="Disabled" disabled />);

    const input = screen.getByRole('textbox');
    expect(input).toBeDisabled();

    await userEvent.type(input, 'Hello World');
    expect(input).toHaveValue('');
  });

  it('should render sub element when endAdornment is provided', () => {
    render(<TextField label="endAdornment" endAdornment={<button>hello btn!</button>} />);

    expect(screen.getByRole('button', { name: 'hello btn!' })).toBeInTheDocument();
  });
});
