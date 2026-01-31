import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import LoadingDots from '../loading-dots';

describe('LoadingDots', () => {
  it('should render three dots', () => {
    const { container } = render(<LoadingDots className="bg-white" />);
    const dots = container.querySelectorAll('.animate-blink');
    expect(dots).toHaveLength(3);
  });

  it('should apply custom className to all dots', () => {
    const { container } = render(<LoadingDots className="bg-blue-500" />);
    const dots = container.querySelectorAll('.bg-blue-500');
    expect(dots).toHaveLength(3);
  });

  it('should have animation delay on second and third dots', () => {
    const { container } = render(<LoadingDots className="bg-white" />);
    const dots = container.querySelectorAll('.animate-blink');

    expect(dots[1]).toHaveClass('animation-delay-[200ms]');
    expect(dots[2]).toHaveClass('animation-delay-[400ms]');
  });

  it('should have inline-flex container', () => {
    const { container } = render(<LoadingDots className="bg-white" />);
    const wrapper = container.querySelector('.inline-flex');
    expect(wrapper).toBeInTheDocument();
  });

  it('should have rounded dots', () => {
    const { container } = render(<LoadingDots className="bg-white" />);
    const dots = container.querySelectorAll('.rounded-md');
    expect(dots).toHaveLength(3);
  });

  it('should have correct dot dimensions', () => {
    const { container } = render(<LoadingDots className="bg-white" />);
    const dots = container.querySelectorAll('.w-1.h-1');
    // Using class check for dimensions
    const allDots = container.querySelectorAll('.w-1');
    expect(allDots).toHaveLength(3);
  });
});
