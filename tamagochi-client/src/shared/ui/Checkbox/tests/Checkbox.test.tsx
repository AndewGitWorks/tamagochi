import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Checkbox } from '../Checkbox';
import { useState } from 'react';

describe('Checkbox', () => {
  it('should render component with label', () => {
    render(<Checkbox label="Approve" />);

    expect(screen.getByRole('checkbox', { name: 'Approve' })).toBeInTheDocument();
  });

  it('should render component with children and label', () => {
    render(
      <Checkbox label="Done">
        <span>Check</span>
      </Checkbox>,
    );
    expect(screen.getByRole('checkbox', { name: /^Done/ })).toBeInTheDocument();
  });

  it('should toggle checked state on click (uncontrolled component)', async () => {
    render(<Checkbox label="Toggle me" />);
    const checkbox = screen.getByRole('checkbox', { name: 'Toggle me' });

    expect(checkbox).not.toBeChecked();

    await userEvent.click(checkbox);
    expect(checkbox).toBeChecked();

    await userEvent.click(checkbox);
    expect(checkbox).not.toBeChecked();
  });

  it('should toggle checked state on click (controlled component)', async () => {
    const TestComponent = () => {
      const [checked, setChecked] = useState(false);

      return (
        <Checkbox
          label="Controlled"
          checked={checked}
          onChange={(e) => setChecked(e.target.checked)}
        />
      );
    };

    render(<TestComponent />);
    const checkbox = screen.getByRole('checkbox', { name: 'Controlled' });

    expect(checkbox).not.toBeChecked();

    await userEvent.click(checkbox);
    expect(checkbox).toBeChecked();
  });

  it('should be disabled when disabled prop is provided', async () => {
    render(<Checkbox label="Disabled click" disabled />);
    const checkbox = screen.getByRole('checkbox', { name: 'Disabled click' });

    expect(checkbox).not.toBeChecked();
    expect(checkbox).toBeDisabled();

    await userEvent.click(checkbox);
    expect(checkbox).not.toBeChecked();
  });
});
